# Feature: fecha límite de tareas, solo día (issue #3)

## Encargo y criterios de aceptación

Encargo literal de la issue #3 de `clerma-arch/desarrollo-agentico-todo-03` (título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"):

- Las tareas pueden tener una fecha límite.
- La fecha es solo día: no tiene hora ni zona horaria visible para el usuario.

Criterios de aceptación (verificables):

1. Al crear una tarea se puede indicar una fecha opcional; sin fecha, la tarea queda sin fecha límite.
2. Al editar una tarea se puede poner, cambiar o quitar la fecha límite (campo vacío = sin fecha).
3. La lista de tareas muestra la fecha límite de cada tarea que la tenga, en formato `DD/MM/AAAA`.
4. La fecha se persiste como día (columna `date`) y se recupera sin hora.
5. Una fecha mal formada no se guarda: se muestra error de validación y se conservan el título y la fecha introducida.
6. Las escrituras (crear, editar, completar, eliminar) siguen protegidas por CSRF.

Requisitos explícitos: solo la fecha límite y "sin hora".
Inferencias (no dichas en la issue, tomadas del estado del código): campo opcional, formato de entrada `YYYY-MM-DD` (control `type="date"`), mensaje de error en español, visualización en `DD/MM/AAAA`.

## Comportamiento actual y propuesta

Estado real de la rama `feat/issue-3-add-task-due-date` (verificado al escribir este plan):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. Existe, con su `.down.fizz`.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. El modelo no valida la fecha; la validación está en la acción.
- `actions/tasks.go`:
  - `parseDueDate` convierte `YYYY-MM-DD` en día sin hora; cadena vacía devuelve `nil`.
  - `TasksCreate` y `TasksUpdate` leen `due_date`, añaden el error `due_date` si el formato es inválido y re-renderizan con 422 conservando la entrada (`dueDateInput`).
  - `TasksEdit` y `TasksIndex` fijan `dueDateInput` a `""` y la fecha de la tarea sale de `task.DueDate`.
  - `TasksToggle` conserva la fecha porque actualiza el mismo registro.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de alta, bloque de error `due_date` y `<span class="task-due">Fecha límite: DD/MM/AAAA</span>` por tarea.
- `templates/tasks/edit.plush.html`: input `type="date"` con ayuda "Opcional. Solo indica el día, sin hora." y error `due_date`.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]` (ancho automático).
- Tests existentes (no hace falta crearlos de nuevo):
  - `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
  - `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksToggle_KeepsDueDate`.

Propuesta: no rediseñar nada. La funcionalidad ya cubre los criterios 1–6 en el código. El trabajo pendiente es **verificar y cerrar**, no implementar. Solo se tocará código si la verificación de la sección de tests falla.

## Alcance

Incluido:
- Confirmar que los criterios de aceptación se cumplen con el código actual.
- Ejecutar las validaciones y registrar su resultado.
- Corregir solo lo que una verificación demuestre roto dentro de la fecha límite.

Excluido:
- Recordatorios, notificaciones, ordenación o filtro por fecha, hora o zona horaria.
- Cambios en la migración ya existente (no reescribir ni renombrar).
- Limpieza de los ~120 archivos `plans/issue-3-*` sin confirmar, de la configuración `config/adw.yml` modificada y de `.claude/skills/ship/`, `.txt`, `.x`, `migrations/schema.sql` (todo esto ya estaba sin confirmar al empezar; no entra en el PR).
- Refactorizar `actions/tasks.go` (la duplicación de `parseDueDate`/validación entre create y update es aceptable para este alcance).

## Diseño por capas

- **Persistencia**: columna `due_date date NULL` en `tasks`. Ya existe. Sin cambios.
- **Modelo**: `Task.DueDate *time.Time`. Ya existe. Sin cambios.
- **Acción y ruta**: `TasksCreate`, `TasksUpdate`, `TasksEdit`, `TasksIndex`. Ya existen. Rutas sin cambios.
- **Renderizado**: dos plantillas Plush y el formato `02/01/2006` para mostrar. Ya existen.
- **CSS**: `.task-due` y el ancho del input de fecha. Ya existen.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date`. |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna. |
| `models/task.go` | Campo `DueDate`. |
| `actions/tasks.go` | `parseDueDate`, alta, edición, listado, toggle. |
| `templates/tasks/index.plush.html` | Campo de alta, error y visualización. |
| `templates/tasks/edit.plush.html` | Campo de edición y error. |
| `public/assets/app.css` | Estilos de `.task-due` y del input de fecha. |
| `models/task_test.go` | Tests de persistencia de la fecha. |
| `actions/tasks_test.go` | Tests de HTTP de la fecha. |

## Implementación paso a paso

No hay pasos de implementación obligatorios. La secuencia es:

1. Confirmar la rama y que la migración está aplicada en el entorno de test.
2. Ejecutar las validaciones de la sección siguiente.
3. Revisar que cada criterio de aceptación tiene al menos un test que lo demuestra (tabla de cobertura en el informe final).
4. Si todo pasa: no hay cambios de código. Pasar al paso de cierre (commit de lo que falte y PR).
5. Si algo falla: parar, registrar la salida y no ampliar el alcance.

## Tests y validación

Condiciones previas:
- PostgreSQL accesible y configurado para `GO_ENV=test` (los tests de `actions` y `models` lo necesitan; `actions/` usa `todo_test`).
- Sin `GO_ENV=test` la app escribe en la base de desarrollo: exportarlo siempre.

Comandos y resultado observado al escribir este plan:

| Comando | Resultado |
| --- | --- |
| `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/` | OK (`todo/actions` y `todo/models`). |
| `go build ./...` | OK. |
| `go vet ./...` | Sin errores (no imprimió salida). |
| `gofmt -l .` | Lista casi todos los `.go` del repo, incluidos los de `main` y de `.claude/worktrees/`. Causa probable: los archivos de trabajo tienen finales de línea CRLF (`w/crlf`, índice `i/lf`) con `core.autocrlf=true`. No es un defecto de la funcionalidad, pero debe confirmarse antes del PR. |
| `GO_ENV=test go test -p 1 ./adws/` | FALLA en `TestTriggerOnceRunsIssuesInSequence`, `TestTriggerShellWrapper`, `TestTriggerConfigurablePromptIsLiteral`, `TestTriggerEmptyListDoesNotInvokeClaude`, `TestTriggerStopsOnClaudeError`, `TestTriggerPollsWithoutRememberingIssues`: `GraphQL: Could not resolve to a Repository with the name 'example/todo'`. No está relacionado con la fecha límite. Pendiente: confirmar si también falla en `main`. |
| `go test -p 1 ./...` completo | Falla solo por `adws` (ver fila anterior). |

Criterio para cerrar: `actions` y `models` en verde, `go build` y `go vet` sin errores, `gofmt` limpio solo en los archivos del cambio, y el fallo de `adws` documentado como preexistente (o corregido fuera de este PR si se confirma que es de `main`).

## Riesgos y recuperación

- **Rama con trabajo acumulado y plans duplicados**: la rama tiene decenas de documentos `plans/issue-3-*` sin commitear y varios commits solo de plans. Riesgo de PR ruidoso. Mitigación: incluir en el PR solo el código y este plan, con lista explícita de archivos.
- **Finales de línea CRLF**: pueden meter diff de línea completo si se reescribe un archivo. Mitigación: no reformatear archivos que no cambian; no ejecutar `gofmt -w` sobre todo el repo.
- **Fallo de `adws` preexistente**: podría bloquear un `go test ./...` completo. Mitigación: documentar, ejecutar `actions` y `models` explícitamente, no ocultar el fallo.
- **Zona horaria**: los tests crean fechas en UTC y el listado formatea el día tal cual. Si la base o el driver convierten la fecha con desfase, la visualización podría mostrar el día anterior. Comprobar con `Test_TasksIndex_ShowsDueDate` y `Test_Task_DueDate_DayOnly_Persists` (ya cubren el caso).
- **Recuperación**: los cambios son aditivos. Revertir la funcionalidad implicaría `migrate down` de `20261022120000` y quitar los campos de plantilla; no se recomienda salvo bloqueo grave.
