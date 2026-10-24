# Feature: fecha límite opcional (solo día) en las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`. Título: "añadir fechaimie alas task" (el título es ambiguo; se lee como "añadir fecha límite a las tareas"). Cuerpo: "vale un dia sin hora".

Requisitos explícitos (del cuerpo de la issue):
- La fecha límite es solo de día, sin hora.

Inferencias (no escritas en la issue):
- La fecha es opcional: una tarea puede no tener fecha.
- Se puede indicar al crear y al editar una tarea.
- La lista muestra la fecha como `DD/MM/YYYY`; la edición la precarga en formato `YYYY-MM-DD`.
- No se valida que la fecha sea futura ni posterior a hoy.

Criterios de aceptación:
1. Crear una tarea sin fecha funciona como antes.
2. Crear una tarea con `2026-10-22` la guarda como día sin hora y la lista la muestra como `22/10/2026`.
3. Editar permite cambiar o borrar la fecha sin alterar el título ni `Completed`.
4. Una fecha mal formada no se guarda: respuesta 422 con el mensaje "La fecha no es válida." en `due_date`, nunca 500.
5. La migración añade una columna `date` nullable; las tareas existentes quedan sin fecha y la migración es reversible.
6. Si el alta falla por validación (p. ej. título vacío), el campo de fecha conserva el valor enviado en vez de vaciarse.

## Comportamiento actual y propuesta

Estado verificado en la rama `feat/issue-3-add-task-due-date` (commits `d13ec2c`, `7913cf0`, `69e8826`, todavía no fusionados en `main`). **La funcionalidad ya está implementada y cubre los criterios 1 a 6.** Este plan no la rediseña: confirma la cobertura, deja explícitas las decisiones y fija la validación que falta ejecutar antes del PR.

Verificado en esta sesión:
- `go build ./...`: correcto.
- `go vet ./...`: sin salida (correcto).
- `go test -count=1 -p 1 ./actions ./models` con PostgreSQL escuchando en `127.0.0.1:5432`: `ok` en ambos paquetes.
- `go test -p 1 ./adws`: falla en `TestTriggerOnceRunsIssuesInSequence`, `TestTriggerShellWrapper`, `TestTriggerConfigurablePromptIsLiteral`, `TestTriggerEmptyListDoesNotInvokeClaude`, `TestTriggerStopsOnClaudeError` y `TestTriggerPollsWithoutRememberingIssues`, con `Could not resolve to a Repository with the name 'example/todo'`. Es un fallo previo, no relacionado: el paquete `adws` no se toca en la rama (`git diff main...HEAD -- adws` vacío) y sus tests dependen de `gh` y de la red.

Implementado en la rama:
- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. El `.down.fizz` hace `drop_column("tasks", "due_date")`.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` no cambia.
- `actions/tasks.go`:
  - `parseDueDate(raw string) (*time.Time, error)`: cadena vacía devuelve `nil`; en otro caso `time.Parse("2006-01-02", raw)`.
  - `TasksCreate` y `TasksUpdate` leen `due_date` como `string`, lo parsean y, si falla, añaden `dueDateInvalidMessage` a `due_date`. Responden 422 re-renderizando la plantilla con `task` (que conserva la fecha enviada cuando es válida).
- `templates/tasks/index.plush.html`: `<input type="date" name="due_date">` en el alta, con `value` tomado de `task.DueDate`; muestra `Fecha límite: DD/MM/YYYY` solo si `t.DueDate != nil`.
- `templates/tasks/edit.plush.html`: `<input type="date" name="due_date">` precargado, con texto de ayuda y mensaje de error.
- `public/assets/app.css`: clase `.task-due` y `.new-task-row input[type="date"]`, con variables existentes.
- Tests presentes:
  - `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
  - `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`.

Cobertura por criterio:
| Criterio | Evidencia |
|---|---|
| 1 | `Test_Task_DueDate_Optional`, `Test_TasksCreate_*` sin fecha existentes |
| 2 | `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_Task_DueDate_DayOnly_Persists` |
| 3 | `Test_TasksUpdate_SetsAndClearsDueDate` |
| 4 | `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate` |
| 5 | Migración `up`/`down` revisadas; no hay test automático de la migración (ver riesgos) |
| 6 | `Test_TasksCreate_InvalidTitleKeepsDueDate` |

## Alcance

Incluido:
- Verificar que la rama cumple los criterios 1 a 6.
- Ejecutar la validación completa en el entorno de test y registrar el resultado.
- Preparar el PR desde esta rama con solo los archivos del cambio y del plan.

Excluido:
- Cambios de código de aplicación, plantillas o CSS: no son necesarios para los criterios.
- Validación de fechas pasadas o futuras, zonas horarias, recordatorios y filtros por fecha: no pedidos.
- Consolidar o borrar los planes previos de la issue #3 en `plans/`: se deja como decisión del usuario (ver incertidumbres).
- Cambios en `adws/`, incluidos los tests que fallan por `example/todo`.
- Cambios no confirmados que ya existían antes de empezar: `config/adw.yml`, `plans/add-task-due-date.md`, `migrations/schema.sql`, `.claude/skills/ship/` y los planes `plans/issue-3-*` no versionados.

## Diseño por capas

- **Persistencia**: una columna `due_date date NULL` en `tasks`. Migración `20261022120000`, posterior a `20260901000000`.
- **Modelo**: `DueDate *time.Time`. El puntero representa "sin fecha" sin valor centinela.
- **Acción**: `parseDueDate` centraliza la conversión `YYYY-MM-DD` → día a medianoche UTC. `TasksCreate` y `TasksUpdate` tratan el error de fecha como error de validación (422), no como 500.
- **Renderizado**: alta y edición usan `<input type="date">`; la lista usa `Format("02/01/2006")`.
- **CSS**: dos reglas nuevas que reutilizan variables existentes (`--color-muted`).
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date`. |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna. |
| `models/task.go` | Campo `DueDate` y etiquetas `db`/`form`. |
| `models/task_test.go` | Tests de modelo: opcional y persistencia solo día. |
| `actions/tasks.go` | `parseDueDate`, validación en create y update, 422 con mensaje. |
| `actions/tasks_test.go` | Tests de acción: alta, edición, lista, errores y precarga. |
| `templates/tasks/index.plush.html` | Campo de fecha en el alta, fecha en la lista, error y conservación del valor. |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición, ayuda y error. |
| `public/assets/app.css` | Estilos `.task-due` y del campo de fecha del alta. |
| `config/adw.yml` | No forma parte del cambio (ver alcance). |

## Implementación paso a paso

No hay implementación pendiente. Los pasos son de verificación y de preparación del PR:

1. Confirmar la rama: `git status --short --branch` debe mostrar `feat/issue-3-add-task-due-date`. Los cambios no confirmados que aparecen en el estado se excluyen del PR.
2. Preparar la base de datos de test como hace CI: `soda migrate -e test` con `GO_ENV=test`, sobre `todo_test` en `127.0.0.1:5432`.
3. Ejecutar la validación de la sección siguiente.
4. Si todo pasa, preparar el PR con los archivos del cambio (migración, modelo, acción, tests, plantillas, CSS) y el plan `plans/issue-3-cumplir-fecha-limite-solo-dia.md`, sin `git add .`.
5. Si algo falla, detenerse y reportar la salida; no ampliar el alcance.

## Tests y validación

Comandos pertinentes a esta funcionalidad:

- `gofmt -l` sobre los archivos Go del cambio. Advertencia de entorno: en esta copia de trabajo los archivos tienen terminaciones CRLF (`core.autocrlf=true`) y `gofmt -l .` lista casi todos los `.go`, aunque el contenido de `HEAD` está limpio. Para comprobarlo sin falsos positivos: `git show HEAD:actions/tasks.go | gofmt -l` (sin salida = correcto), o `gofmt -d` y revisar que la única diferencia es el fin de línea.
- `go vet ./...`: ya verificado, sin salida.
- `go build ./...`: ya verificado, correcto.
- `GO_ENV=test go test -count=1 -p 1 ./actions ./models`: ya verificado, `ok`. Debe repetirse tras `soda migrate -e test`.
- `go test -p 1 ./adws`: fallo previo no relacionado; se reporta, no se corrige aquí.

Criterio de paso: `actions` y `models` en `ok`; `go vet` y `go build` sin salida; `gofmt` sin diferencias de contenido.

## Riesgos y recuperación

- **Migración no probada automáticamente**: el criterio 5 depende de revisión manual. Recuperación: `soda migrate down` aplica el `.down.fizz`; si la columna ya existe en otra base, la migración fallaría y habría que revisar el estado de `schema_migration`.
- **Zona horaria**: la fecha se guarda como día a medianoche UTC. Los tests de persistencia pasan contra PostgreSQL local; una base con otra configuración de zona podría desplazar el día. Mitigación: los tests comparan año, mes, día y hora sobre el valor recargado.
- **Planes duplicados**: `plans/` contiene ya varios planes de la issue #3 (`add-task-due-date.md`, `issue-3-fecha-limite-solo-dia.md`, `issue-3-verificar-fecha-limite.md`, entre otros). No afectan al código, pero confunden. No se borran aquí.
- **Cambios no confirmados fuera de alcance**: `config/adw.yml` cambia el prompt del trigger a `/ship`. Si se confirma con el PR, mezclaría configuración de automatización con la feature. Se excluye.
- **Fallo previo de `adws`**: si CI ejecuta `go test ./...` completo, fallará por los tests de `adws` con independencia de este cambio.

## Incertidumbres pendientes

- ¿Se quiere mantener la fecha pasada permitida? Se asume que sí (no se pidió validarla).
- ¿Se quiere un test de regresión para `TasksUpdate` con título vacío que conserve la fecha? No se pidió; no se añade.
- ¿Qué plan de la issue #3 se considera canónico? Decisión del usuario.
