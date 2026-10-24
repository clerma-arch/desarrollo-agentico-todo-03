# Feature: Fecha límite opcional de tarea (solo día) — verificación del estado y cierre (issue #3)

## Encargo y criterios de aceptación

Encargo de la issue #3 (`añadir fechaimie alas task`, cuerpo: `vale un dia sin hora`):

- Explícito: las tareas deben poder tener una fecha (`fecha límite`) añadida a la lista de tareas.
- Explícito: la fecha es solo de día, sin hora.

Criterios de aceptación:

1. Crear una tarea con fecha opcional: si se envía `due_date` en formato `YYYY-MM-DD`, la tarea se guarda con ese día.
2. Crear una tarea sin fecha: `due_date` queda vacío (NULL) y la tarea funciona como antes.
3. Una fecha con formato inválido (p. ej. `2026-13-40`) no crea la tarea y muestra el error `La fecha no es válida.` en el campo `due_date`.
4. Editar una tarea permite fijar, cambiar y borrar la fecha, y el formulario de edición precarga la fecha existente.
5. Completar/reabrir una tarea conserva su fecha.
6. La lista muestra la fecha con formato `dd/mm/aaaa` solo cuando existe.
7. Las acciones que cambian estado (crear, actualizar) siguen protegidas por CSRF.

Inferencias (no explícitas en la issue): nombre del campo `due_date`, etiqueta "Fecha límite", campo nullable, mensaje de error y uso de `input type="date"`. Son las que ya aparecen en la implementación del branch.

## Comportamiento actual y propuesta

Estado real verificado en el branch `feat/issue-3-add-task-due-date` (HEAD `d18bc2c`, PR #5 abierto desde esta rama):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: añade `due_date` tipo `date`, nullable. Ya commiteada.
- Modelo `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` no cambia (solo valida el título).
- Acción `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en `*time.Time`; `TasksCreate` y `TasksUpdate` usan `task.Validate(tx)` y añaden el error `due_date` si la fecha no se parsea, luego `tx.Create` / `tx.Update`. `TasksIndex` y `TasksEdit` inicializan `dueDateInput`.
- Plantillas: `templates/tasks/index.plush.html` (input date en el formulario de alta, fecha mostrada en la lista) y `templates/tasks/edit.plush.html` (input date con precarga y error).
- Estilos: `public/assets/app.css` (clases `.task-due`, `.field-hint` y similares).
- Tests: `actions/tasks_test.go` y `actions/actions_test.go` (helper `authed` con token CSRF), `models/task_test.go` (persistencia día sin hora y campo opcional).

Propuesta: **no reimplementar**. Verificar que el código del branch cumple los criterios 1–7, ejecutar las validaciones y dejar el PR #5 listo. Cualquier cambio de código debe ser mínimo y solo si una verificación falla.

## Alcance

Incluido:
- Verificación de criterios de aceptación contra el código y los tests existentes.
- Ejecución de validaciones (`gofmt`, `go vet`, `go build`, `go test`).
- Actualización del PR #5 (push a `origin` de la rama) si no hay cambios de código, o commit mínimo si lo hay.

Excluido:
- Hora de la fecha límite, recordatorios, ordenación por fecha, filtros.
- Cambios en `adws/`, `config/adw.yml`, `migrations/schema.sql`, `.txt`, `.x`, `.claude/skills/ship/` y los `plans/*.md` sin confirmar (no forman parte de este cambio).
- Reformatear archivos a LF (ver riesgos).

## Diseño por capas

- **Persistencia**: migración `add_column("tasks", "due_date", "date", {"null": true})` (ya existe). Usa tipo `date`, por eso el requisito "sin hora" queda garantizado por la base de datos.
- **Modelo**: `*time.Time` nullable; sin lógica adicional.
- **Acción**: parseo y validación de la entrada en la acción (como el resto de validaciones de formulario del proyecto con `verrs.Add`).
- **Renderizado**: Plush con `input type="date"`, precarga y errores en el mismo patrón de `title`.
- **JavaScript**: ninguno.
- **CSS**: clases auxiliares en `public/assets/app.css` ya presentes.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date` nullable |
| `models/task.go` | Campo `DueDate` |
| `actions/tasks.go` | `parseDueDate`, create/update con fecha, precarga |
| `templates/tasks/index.plush.html` | Campo de fecha en alta y fecha en lista |
| `templates/tasks/edit.plush.html` | Campo de fecha en edición |
| `actions/tasks_test.go` | Tests de comportamiento HTTP |
| `actions/actions_test.go` | Helper CSRF `authed` |
| `models/task_test.go` | Persistencia del día sin hora |

## Implementación paso a paso

1. Confirmar que `git status` solo muestra los archivos no relacionados ya conocidos (`config/adw.yml`, `.claude/skills/ship/`, `.txt`, `.x`, `migrations/schema.sql`, `plans/*.md` sin confirmar). No stagearlos.
2. Revisar contra los criterios 1–7 el diff `git diff main...HEAD -- actions models templates migrations public`. Si todos se cumplen, no hay cambios de código.
3. Ejecutar las validaciones de la sección siguiente. Si alguna falla por código de este cambio, detenerse y corregir solo el fallo, sin ampliar alcance.
4. Si no hay cambios de código, añadir únicamente este plan al branch (commit `docs(plans): …`) y hacer `git push origin feat/issue-3-add-task-due-date` (sin `--force`). El PR #5 se actualiza solo.

## Tests y validación

Entorno: PostgreSQL disponible y migraciones de test preparadas. Ejecutar siempre con `GO_ENV=test` (ver memoria del proyecto: sin esta variable la app escribe en `development`).

Comandos y resultado observado el 2026-10-24:

- `go vet ./...` → sin salida (OK).
- `go build ./...` → sin salida (OK).
- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/` → `ok todo/actions`, `ok todo/models`.
- `GO_ENV=test go test -p 1 ./...` → `adws` falla (ver riesgos); `actions` y `models` OK.
- `gofmt -l .` → lista casi todos los archivos del repositorio, también los de `main`. Es un problema de finales de línea CRLF (`file models/task.go`: "with CRLF line terminators"), no de formato Go. No reformatear.

Tests que demuestran el comportamiento (ya existen):
- `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`.
- `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate`.
- `models`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.

Un test de regresión confirmado: los tests de creación sin fecha (`as.Nil(task.DueDate)`, línea 41) y de fecha inválida cubren el criterio 2 y 3.

## Riesgos y recuperación

- **Fallo preexistente en `adws`**: `TestTrigger*` ejecuta `gh issue list --repo example/todo` (llamada real, sin mock efectivo) y falla con `Could not resolve to a Repository`. El paquete `adws/` no aparece en `git diff main...HEAD`, así que no lo causa este cambio. No se ha verificado en `main` directamente; comprobarlo antes de declararlo preexistente. No es un bloqueo para este cambio, pero el PR no debe presentarse como "todo en verde" sin mencionarlo.
- **`gofmt -l`**: CRLF en todo el repo. Reformatear cambiaría todas las líneas. Dejarlo tal cual y mencionarlo en el PR.
- **Cambio sin commit en `config/adw.yml`** (repo `clerma-arch/...` y prompt `/ship`): fuera del PR. No commitear.
- **Zona horaria**: `time.Parse` devuelve UTC; con columna `date` se guarda el día. Verificar en la vista de lista que el día mostrado es el introducido (tests `Test_TasksIndex_ShowsDueDate` y `Test_Task_DueDate_DayOnly_Persists`).
- **Recuperación**: si hay que deshacer el cambio de la migración, usar el `.down.fizz` (`drop_column("tasks", "due_date")`) en entorno de test; nunca en desarrollo sin copia.

## Decisiones y pendientes

- Decisión: reutilizar la implementación existente del branch en lugar de planificar una nueva.
- Pendiente: confirmar en `main` si `adws` ya falla antes de este branch.
- Pendiente: decidir si se normalizan los finales de línea del repositorio en un cambio aparte (fuera de este PR).
