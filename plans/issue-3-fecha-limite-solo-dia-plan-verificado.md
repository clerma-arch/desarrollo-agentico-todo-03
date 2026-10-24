# Feature: Fecha límite opcional de tareas (solo día, sin hora)

## Encargo y criterios de aceptación

Fuente: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título «añadir fechaimie alas task» (leído con `gh issue view 3`). Se interpreta «fechaimie» como «fecha límite». El cuerpo de la issue dice «vale un dia sin hora».

Requisitos explícitos:

1. Cada tarea puede tener una fecha límite.
2. La fecha límite es solo un día, sin hora.

Criterios de aceptación observables (inferidos de los requisitos y del estado actual de la rama):

- Al crear una tarea, el formulario tiene un campo de fecha opcional. Sin fecha, la tarea se crea como hasta ahora.
- Al crear o editar una tarea con fecha válida (`YYYY-MM-DD`), la fecha se guarda como día, sin hora.
- En la lista, las tareas con fecha muestran «Fecha límite: dd/mm/aaaa». Las tareas sin fecha no muestran nada.
- En el formulario de edición, el campo aparece precargado con la fecha guardada.
- Una fecha no válida se rechaza con el mensaje «La fecha no es válida.», sin perder el título ni la fecha introducida.
- Completar, reabrir o editar solo el título no borra la fecha límite.
- Vaciar el campo de fecha en la edición borra la fecha límite.

Inferencias (no están en la issue, quedan documentadas para revisión):

- El formato de visualización `dd/mm/aaaa` es una convención del repositorio para textos en español; la issue no lo fija.
- «Opcional» es una inferencia: la issue no dice que sea obligatoria, y las tareas existentes no tienen fecha.
- Editar una tarea envía siempre el campo `due_date` desde el formulario, así que un envío sin ese campo borraría la fecha. Es el comportamiento actual; los tests de acción no cubren un PUT sin el campo.

## Comportamiento actual y propuesta

Estado verificado en la rama `feat/issue-3-add-task-due-date` (commit `2c56d65`), diferencias respecto a `main`:

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: añade la columna `due_date` `date`, nullable. El `down` la elimina.
- `models/task.go`: el campo `DueDate *time.Time` con etiquetas `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en un `time.Time` sin hora; cadena vacía significa sin fecha; formato inválido añade un error `due_date`. `TasksCreate` y `TasksUpdate` lo usan y re-renderizan con código 422 si hay errores. `TasksIndex`, `TasksEdit` y `TasksToggle` ya trabajan con el modelo.
- `templates/tasks/index.plush.html`: campo `type="date"` junto al título y muestra de la fecha en la lista.
- `templates/tasks/edit.plush.html`: campo `type="date"` con ayuda «Opcional. Solo indica el día, sin hora.» y errores.
- `public/assets/app.css`: clases `.task-due` y ajuste de ancho del campo de fecha en `.new-task-row`.

Propuesta: no hay cambios de código pendientes. Quedan validar, revisar los archivos que deben entrar en el PR y abrirlo. Si la validación revela un fallo, se corrige sin ampliar el alcance.

## Alcance

Incluido (ya presente en la rama):

- Migración, modelo, acciones create/edit/update y vistas de índice y edición.
- Tests de modelo y de acción para fecha opcional, día sin hora, fecha inválida, conservación de la fecha y precarga.
- Estilos mínimos del campo y de la fecha mostrada.

Excluido:

- Hora, zona horaria, recordatorios, filtros o ordenación por fecha.
- Cambios en `adws/`, `config/adw.yml` y otros archivos ajenos a la issue.
- Cualquier cambio de esquema más allá de la columna `due_date`.

## Diseño por capas

- **Persistencia**: columna `due_date date NULL` en `tasks`, creada por la migración de la fecha. Un tipo `date` evita guardar hora.
- **Modelo**: `Task.DueDate *time.Time` nullable. La validación de título no cambia.
- **Acción y ruta**: sin rutas nuevas. `POST /tasks`, `GET /tasks/{task_id}/edit` y `PUT /tasks/{task_id}` reciben `due_date` como campo de formulario.
- **Renderizado**: campos `type="date"` y texto de fecha con `Format("02/01/2006")`.
- **CSS**: solo clases de presentación; no hay JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añadir y quitar la columna `due_date`. |
| `models/task.go` | Campo `DueDate` y validación del título. |
| `actions/tasks.go` | `parseDueDate`, validación de fecha en create y update. |
| `templates/tasks/index.plush.html` | Campo de fecha en el alta y fecha mostrada. |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición, ayuda y errores. |
| `public/assets/app.css` | Estilos `.task-due` y `.new-task-row input[type="date"]`. |
| `models/task_test.go` | `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`. |
| `actions/tasks_test.go` | Tests de alta, edición, índice y precarga con fecha. |
| `actions/actions_test.go` | Suite de acciones (`Test_ActionSuite`). |

## Implementación paso a paso

Como el código ya existe, la implementación consiste en confirmar el estado y cerrar el PR:

1. Confirmar que la rama contiene los cambios de la tabla anterior y que `git diff main...HEAD --stat` coincide con ellos.
2. Ejecutar las validaciones de la sección siguiente y revisar que todas pasen.
3. Preparar el PR: incluir solo los archivos de la tabla y este plan. Excluir lo que no es de la issue (ver Riesgos).
4. Hacer push de la rama a `origin` y abrir el PR contra `main`.

## Tests y validación

Estado comprobado al planificar:

- `go vet ./...`: sin salida, correcto.
- `go build ./...`: correcto.
- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/`: `ok todo/actions` y `ok todo/models`.
- `gofmt -l .`: lista muchos archivos, pero solo por finales de línea CRLF del checkout de Windows (`core.autocrlf=true`). Los blobs de `HEAD` y el contenido con LF pasan `gofmt -l`. No es un fallo de código.

Requisitos para ejecutar los tests:

- PostgreSQL escuchando en `127.0.0.1:5432` con la base `todo_test` y las migraciones aplicadas.
- `GO_ENV=test` obligatorio para `actions/`. Sin él, la app escribe en `todo_development` y la suite lee `todo_test`, lo que produce fallos en cascada no relacionados con el código.

Comandos de validación del PR:

- `gofmt -l .`: debe listar solo archivos con CRLF; comprobar el contenido con `tr -d '\r' < archivo | gofmt -l`.
- `go vet ./...`
- `go build ./...`
- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/`
- `go test -p 1 ./...` completo: los tests de `adws/` fallan en Windows por el `gh` real y también fallan en `main`; no son parte de esta feature. No se ha ejecutado en esta sesión.

## Riesgos y recuperación

- **Cambios ajenos en el working tree**: antes de empezar había cambios sin confirmar que no pertenecen a la issue: `M config/adw.yml`, `?? .claude/skills/ship/`, `?? .txt`, `?? .x`, `?? migrations/schema.sql`, y decenas de `plans/*.md` de intentos anteriores. No deben entrar en el PR. Los `plans/` de esta issue tampoco hace falta subirlos todos; se sugiere subir solo este plan.
- **Borrado accidental de la fecha**: el formulario de edición siempre envía `due_date`. Si una petición externa omite el campo, la fecha se borra. Riesgo bajo; no hay API pública documentada para ello.
- **Zona horaria**: `time.Parse` sin zona devuelve UTC medianoche. Como la columna es `date`, el día no cambia en la lectura; los tests de persistencia lo cubren.
- **Recuperación**: si el PR se rechaza, revertir los commits de la rama no afecta a `main`. La migración tiene `down` que elimina la columna.
- **Incertidumbre pendiente**: confirmar con el autor de la issue el formato `dd/mm/aaaa` y que la fecha sea opcional.
