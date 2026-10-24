# Feature: fecha límite opcional (solo día) en las tareas — issue #3

## Encargo y criterios de aceptación

Encargo (issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título «añadir fechaimie alas task», cuerpo «vale un dia sin hora»):

- Las tareas pueden tener una fecha límite.
- La fecha límite es solo día, sin hora.

Criterios de aceptación:

1. Crear una tarea sin fecha límite la deja sin fecha (`NULL`).
2. Crear una tarea con fecha límite válida (`YYYY-MM-DD`) la persiste como día, sin hora.
3. Editar una tarea permite fijar, cambiar y borrar la fecha límite, sin tocar `completed` ni `created_at`.
4. Una fecha inválida devuelve 422 con el mensaje «La fecha no es válida.», conserva el valor introducido y no crea ni modifica la tarea.
5. Un error de título conserva la fecha introducida en el formulario.
6. La lista muestra la fecha como `dd/mm/aaaa` solo cuando existe.
7. Completar o reabrir una tarea no altera su fecha límite.

Requisitos explícitos (de la issue): fecha límite en la tarea; solo día, sin hora.

Inferencias (no explícitas en la issue; ya reflejadas en la rama):

- La fecha es opcional, porque la issue no pide que sea obligatoria y el resto de tareas existentes no la tienen.
- El formato de entrada es el de `<input type="date">` (`YYYY-MM-DD`).
- La visualización en lista usa `dd/mm/aaaa`, el formato habitual en español.

## Comportamiento actual y propuesta

Estado de la rama `feat/issue-3-add-task-due-date` (ya hecho en commits previos, verificado leyendo el diff contra `main`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: añade `due_date` como `date` nullable; el `down` la elimina.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"`.
- `actions/tasks.go`: `parseDueDate` convierte la cadena vacía en `nil` y el formato `YYYY-MM-DD` en `time.Time`; `TasksCreate` y `TasksUpdate` leen `due_date` desde un struct de entrada; un error de formato se añade a `verrs` bajo la clave `due_date`.
- `templates/tasks/index.plush.html`: campo `due_date` en el formulario de alta, bloque `task-due` con la fecha en la lista.
- `templates/tasks/edit.plush.html`: campo `due_date` con ayuda y error.
- `public/assets/app.css`: clases `.task-due` y ajuste de `input[type="date"]` en `.new-task-row`.
- Tests añadidos en `actions/tasks_test.go`, `actions/actions_test.go` (helper `authed` con token CSRF) y `models/task_test.go`.

Lo que NO está hecho: ninguna validación se ha ejecutado en esta sesión. No hay evidencia de que compile, pase `gofmt`/`go vet` ni de que los tests pasen con la base de prueba.

Propuesta: no reescribir la funcionalidad. Verificarla con las validaciones, corregir solo fallos que aparezcan dentro de los archivos de la feature, y preparar el PR.

## Alcance

Incluido:

- Verificación de la implementación existente (migración, modelo, acciones, plantillas, CSS, tests).
- Corrección de fallos detectados al ejecutar las validaciones, solo en archivos de esta feature.
- Preparación del PR desde esta rama contra `main`.

Excluido:

- Hora de la fecha límite, recordatorios, notificaciones, filtros por fecha u orden por fecha.
- Fecha límite obligatoria o validación de fechas pasadas.
- Cambios en `config/adw.yml`, `migrations/schema.sql`, `.claude/skills/ship/`, `.txt`, `.x` ni en los planes de otras sesiones (ver «Riesgos»).
- Reescritura de la migración ya presente o de los tests que ya cubren el comportamiento.

## Diseño por capas

- **Persistencia y migraciones**: ya existe. Columna `due_date date NULL`. No se añade índice (no hay consultas por fecha).
- **Modelo**: ya existe. `*time.Time` para representar la ausencia de fecha.
- **Acción y ruta**: ya existen `TasksCreate` y `TasksUpdate`. Las rutas no cambian.
- **Renderizado**: ya existe en las dos plantillas. La entrada se re-muestra con `dueDateInput` cuando hay error.
- **JavaScript**: no aplica.
- **CSS**: ya existe `.task-due` y el ajuste del `input[type="date"]`.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date`. |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna. |
| `models/task.go` | Campo `DueDate` y validación del título. |
| `actions/tasks.go` | `parseDueDate`, alta y edición con fecha. |
| `actions/tasks_test.go` | Tests de alta, edición, errores y visualización. |
| `actions/actions_test.go` | Helper `authed` y `csrfToken` para peticiones que cambian estado. |
| `models/task_test.go` | Tests de persistencia de la fecha solo día. |
| `templates/tasks/index.plush.html` | Campo en el alta y fecha en la lista. |
| `templates/tasks/edit.plush.html` | Campo en la edición. |
| `public/assets/app.css` | Estilos de la fecha y del campo de alta. |

## Implementación paso a paso

1. Confirmar que el árbol de la rama contiene los archivos de la tabla y que `git diff main...HEAD` coincide con lo descrito arriba.
2. Ejecutar `gofmt -l .`. Si lista archivos de la feature, formatearlos y repetir.
3. Ejecutar `go vet ./...` y `go build ./...`. Corregir solo errores dentro de los archivos de la feature.
4. Preparar la base de pruebas con las migraciones y ejecutar los tests con `GO_ENV=test` (ver «Tests y validación»).
5. Si algún test falla, corregir el código de la feature (no el test) salvo que el test esté mal planteado respecto a los criterios de aceptación; anotar el motivo.
6. Repetir `gofmt`, `go vet`, `go build` y `go test` hasta que todo pase.
7. No añadir nuevos tests salvo que alguno de los criterios 1–7 no esté cubierto (ver mapeo abajo).

Mapeo criterio → test existente:

- 1 → `models.Test_Task_DueDate_Optional`, `actions.Test_TasksCreate_Valid` (`as.Nil(task.DueDate)`).
- 2 → `models.Test_Task_DueDate_DayOnly_Persists`, `actions.Test_TasksCreate_WithDueDate`.
- 3 → `actions.Test_TasksUpdate_SetsAndClearsDueDate`, `actions.Test_TasksEdit_PrefillsDueDate`.
- 4 → `actions.Test_TasksCreate_InvalidDueDate`, `actions.Test_TasksUpdate_InvalidDueDate`.
- 5 → `actions.Test_TasksCreate_InvalidTitleKeepsDueDate`, `actions.Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- 6 → `actions.Test_TasksIndex_ShowsDueDate`.
- 7 → `actions.Test_TasksToggle_KeepsDueDate`.

## Tests y validación

Condiciones para ejecutar:

- PostgreSQL accesible con la configuración de `database.yml` para el entorno `test`.
- Migraciones aplicadas en la base de prueba (incluida `20261022120000`).
- Variable de entorno `GO_ENV=test`. Sin ella, la aplicación escribe en la base de `development` (ver memoria «go-tests-need-go-env-test»).

Comandos, en este orden:

```
gofmt -l .
go vet ./...
go build ./...
GO_ENV=test go test -p 1 ./...
```

Resultado esperado: `gofmt -l .` sin salida; `go vet` y `go build` sin errores; todos los paquetes de `go test` en `ok`.

Si PostgreSQL no está disponible, los tests de `actions` y `models` no pueden ejecutarse. En ese caso, la validación queda incompleta y debe reportarse como tal; no abrir el PR como verificado.

## Riesgos y recuperación

- **Validaciones nunca ejecutadas.** Riesgo principal: la rama podría no compilar o fallar en tests. Mitigación: los pasos 2–6. Recuperación: corregir en la feature y volver a validar.
- **Entorno de pruebas.** Sin PostgreSQL o sin `GO_ENV=test`, no hay evidencia. Recuperación: preparar el entorno y repetir; no marcar como verificado sin ejecutar.
- **PR ya abierto.** Ya existe el PR #5 (`feat(tasks): añadir fecha límite opcional (solo día) a las tareas`) desde esta rama. Un `gh pr create` duplicaría el PR; el paso de PR debe actualizar el existente (push) y no crear otro.
- **Archivos ajenos a la feature.** Hay cambios sin confirmar fuera del alcance: `M config/adw.yml`, `?? .claude/skills/ship/`, `?? migrations/schema.sql`, `?? .txt`, `?? .x`, y decenas de planes sin commit en `plans/`. No deben entrar en el PR. Hacer stage por lista explícita.
- **Redundancia menor.** `models/task.go` mantiene `form:"due_date"` aunque las acciones lean la fecha desde un struct de entrada. Inofensivo; no cambiarlo salvo que un test lo exija.
- **Formato de la migración.** La marca `20261022120000` es anterior a la fecha de hoy (2026-10-24) y posterior a la de creación (`20260901`). Sin conflicto conocido; no renombrar una migración ya presente en la rama.
- **Rollback.** Revertir con `down`: `drop_column("tasks", "due_date")`. Pérdida de datos: las fechas guardadas se pierden en el rollback.

Incertidumbres pendientes:

- Si la base de pruebas está disponible en este entorno (no se ha comprobado).
- Si el PR #5 debe actualizarse con este cambio o mantenerse tal cual, según la decisión de la persona que lleva el flujo.
