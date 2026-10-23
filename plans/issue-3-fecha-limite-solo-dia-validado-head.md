# Feature: fecha límite de tareas solo día (issue #3)

## Encargo y criterios de aceptación

Encargo de la issue #3 "añadir fechaimie alas task", con el comentario "vale un dia sin hora":

- Cada tarea puede tener una fecha límite opcional.
- La fecha es solo día (sin hora).
- La fecha se puede indicar al crear una tarea y cambiar al editarla.
- Una fecha inválida no se guarda y se muestra un error; el valor introducido se conserva en el formulario.
- La fecha se muestra en la lista de tareas.

Requisitos explícitos (del encargo): fecha límite opcional, solo día.

Inferencias (no escritas en la issue): formato de entrada `YYYY-MM-DD` del `<input type="date">`, visualización `DD/MM/YYYY` en la lista, mensaje de error "La fecha no es válida.", y que completar una tarea no altera la fecha.

## Comportamiento actual y propuesta

Estado observado en `feat/issue-3-add-task-due-date` (HEAD `7a6f8b8`, sincronizada con `origin`). El código committeado ya implementa la funcionalidad:

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `.down.fizz`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` (vacío = sin fecha; formato `2006-01-02`), `TasksCreate` y `TasksUpdate` leen `due_date` del formulario, validan y guardan; en error devuelven 422 con `dueDateInput` para conservar el valor; `TasksEdit` y `TasksIndex` inicializan `dueDateInput`.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de creación, fecha mostrada como `Fecha límite: DD/MM/YYYY` en cada tarea.
- `templates/tasks/edit.plush.html`: campo "Fecha límite" con ayuda "Opcional. Solo indica el día, sin hora." y error asociado.
- `public/assets/app.css`: clase `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Pruebas committeadas: `models/task_test.go` (opcional, persiste solo día) y `actions/tasks_test.go` (crear con/sin fecha, fecha inválida, conservar fecha en errores, mostrar en índice, editar/limpiar, completar no borra la fecha).

Propuesta: no hay cambio de código pendiente. La implementación committeada cumple los criterios. La acción es validar el HEAD committeado y abrir el PR con los commits de la rama, sin incluir los cambios sin confirmar.

## Alcance

Incluido:

- Validar el HEAD committeado con las pruebas y validaciones de este plan.
- Abrir el PR desde `feat/issue-3-add-task-due-date` hacia `main`, con los commits ya existentes de la rama.

Excluido:

- Cualquier cambio en código de aplicación, tests, migraciones, templates o CSS.
- Los cambios sin confirmar: `actions/actions_test.go` (helpers CSRF, no necesarios según el commit `7a6f8b8`), `actions/tasks_test.go` (su versión sin confirmar sustituye `as.HTML` por `as.authed` en tests existentes, cambio fuera de alcance) y `config/adw.yml` (configuración local del trigger).
- Artefactos sin rastrear: `.claude/skills/ship/`, `.txt`, `migrations/schema.sql` y los `plans/issue-3-*.md` previos (salvo este plan, que se añade al PR).
- Los fallos de `adws` (ver Tests y validación): no están relacionados con esta funcionalidad.

## Diseño por capas

Solo capas ya implementadas; no se requiere cambio.

- Persistencia y migraciones: columna `due_date date null`. Ya existe.
- Modelo: `Task.DueDate *time.Time`. Ya existe.
- Acción y ruta: `TasksCreate`, `TasksUpdate`, `TasksEdit`, `TasksIndex`. Ya existen.
- Renderizado: `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`. Ya existen.
- CSS: `public/assets/app.css`. Ya existe.
- JavaScript: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date` |
| `models/task.go` | Campo `DueDate` del modelo |
| `actions/tasks.go` | Parseo, validación y persistencia de la fecha |
| `templates/tasks/index.plush.html` | Campo de fecha al crear y fecha en la lista |
| `templates/tasks/edit.plush.html` | Campo de fecha al editar |
| `public/assets/app.css` | Estilos de la fecha y del input |
| `models/task_test.go` | Pruebas del modelo |
| `actions/tasks_test.go` | Pruebas de acciones y render (committeadas) |

## Implementación paso a paso

1. Confirmar que la rama está limpia en lo committeado: `git status --short --branch` debe mostrar `feat/issue-3-add-task-due-date...origin/feat/issue-3-add-task-due-date` sin commits pendientes de push.
2. Ejecutar las validaciones de la sección siguiente sobre el HEAD committeado. Si alguna falla, detenerse y no abrir el PR.
3. Stage explícito solo de los archivos del cambio (ninguno nuevo de código) y de este plan: `plans/issue-3-fecha-limite-solo-dia-validado-head.md`. No usar `git add .` ni `git add -A`.
4. Si el plan se añade, un commit Conventional Commits: `docs(plans): registrar la validación de la fecha límite solo día (issue #3)`, con `Co-Authored-By: Claude Haiku 5.5 <noreply@anthropic.com>`.
5. Push a `origin` de la rama sin `--force`.
6. Abrir el PR con `gh pr create --base main`, título Conventional Commits (`feat(tasks): añadir fecha límite solo día a las tareas (issue #3)`), y cuerpo que incluya resumen, ruta de este plan, pruebas existentes y resultado de la validación. Terminar con `🤖 Generated with [Claude Code](https://claude.com/claude-code)`.

## Tests y validación

Requisito de entorno: PostgreSQL accesible en `127.0.0.1:5432` (verificado) y `GO_ENV=test`. Sin `GO_ENV=test` las pruebas de aplicación pueden escribir en la base de desarrollo.

Comandos y resultado observados en el HEAD `7a6f8b8`, en un worktree limpio:

- `GO_ENV=test go test -p 1 ./models/ ./actions/`: `ok todo/models`, `ok todo/actions`. Es la validación que demuestra el comportamiento.
- `go vet ./...`: sin salida (correcto).
- `go build ./...`: correcto.
- `gofmt -l` sobre los blobs LF de HEAD (`models/task.go`, `models/task_test.go`, `actions/tasks.go`, `actions/tasks_test.go`): sin salida (correcto). Un `gofmt -l .` sobre el checkout de Windows lista casi todos los archivos porque `core.autocrlf=true` introduce CRLF; no es un problema del código.
- `GO_ENV=test go test -p 1 ./...`: `todo/actions` y `todo/models` pasan; `todo/adws` falla en `TestTrigger*` porque consulta GitHub con el repositorio `example/todo` (`Could not resolve to a Repository`). Es un fallo previo al trabajo de esta feature, presente también en HEAD, y no se corrige en este cambio.

Pruebas que demuestran cada criterio (ya committeadas):

- Fecha opcional y solo día: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_Valid` (fecha nula).
- Crear con fecha: `Test_TasksCreate_WithDueDate`.
- Fecha inválida no se guarda y se conserva: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Conservar fecha al fallar otro campo: `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- Mostrar en la lista: `Test_TasksIndex_ShowsDueDate`.
- Editar, limpiar y prellenar: `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`.
- Completar no borra la fecha: `Test_TasksToggle_KeepsDueDate`.

No se añaden pruebas nuevas: no hay comportamiento pendiente que demostrar.

## Riesgos y recuperación

- Riesgo: que el PR incluya cambios sin confirmar (helpers CSRF, `config/adw.yml`). Mitigación: stage con lista explícita; no tocar esos archivos.
- Riesgo: que `adws` falle en CI y bloquee el PR. Es un fallo previo, independiente de la feature; hay que señalarlo en la descripción del PR.
- Riesgo: la migración `20261022120000` debe ejecutarse antes de los tests en una base limpia. Recuperación: `buffalo test` o `pop migrate` sobre `todo_test`; el down es `drop_column("tasks", "due_date")`.
- Recuperación de un PR erróneo: cerrarlo sin borrar la rama; no usar `--force`.
