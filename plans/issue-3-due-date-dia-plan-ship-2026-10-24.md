# Feature: fecha límite de las tareas, solo día (issue #3)

## Encargo y criterios de aceptación

Fuente: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task" (sic), cuerpo "vale un dia sin hora". Rama `feat/issue-3-add-task-due-date`, base `main`.

**Requisitos explícitos (de la issue):**

- Las tareas pueden tener una fecha límite.
- La fecha es solo día, sin hora.

**Inferencias (no están en la issue):**

- La fecha límite es opcional: una tarea sin fecha sigue siendo válida. Motivo: la issue dice "añadir", no "obligatorio", y las tareas existentes no tienen fecha.
- Entrada: `input type="date"` en formulario, valor `YYYY-MM-DD`.
- Visualización: `dd/mm/aaaa`, coherente con un usuario en español.
- Sin zona horaria ni recordatorios.

**Criterios de aceptación:**

1. Crear una tarea sin fecha funciona como antes y deja `due_date` a NULL.
2. Crear una tarea con fecha guarda solo el día y el listado muestra `Fecha límite: dd/mm/aaaa`.
3. Editar permite fijar, cambiar y borrar la fecha; el formulario de edición muestra la fecha actual.
4. Fecha con formato inválido: 422 con "La fecha no es válida.", no crea ni modifica nada y conserva lo escrito.
5. Título vacío con fecha válida: 422 y conserva la fecha en el formulario.
6. Completar o reabrir una tarea conserva su fecha límite.

## Comportamiento actual y propuesta

**Estado real de la rama (leído del código; los tests no se han ejecutado en esta fase):**

La implementación ya existe en la rama. Respecto a `main` hay 10 archivos de código/plantilla/tests cambiados (272 inserciones, 31 eliminaciones) más `plans/`:

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`; `down.fizz` con `drop_column`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` (vacío → `nil`; si no, `time.Parse("2006-01-02")`); `TasksCreate` y `TasksUpdate` leen `title` y `due_date` con struct de entrada, llaman `task.Validate(tx)`, añaden error `due_date` si no parsea y después `tx.Create` / `tx.Update`. `TasksToggle` no se toca.
- `templates/tasks/index.plush.html`: input `type="date"` en el alta, `<span class="task-due">` con `02/01/2006`, error `due_date`.
- `templates/tasks/edit.plush.html`: input `type="date"` con prefill, ayuda y error `due_date`.
- `public/assets/app.css`: `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Tests: `actions/tasks_test.go` (alta con/sin fecha, fecha inválida, título inválido con fecha conservada, listado, edición fija/borra/inválida, prefill, toggle conserva fecha) y `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`).

**Qué falta:** no se ha identificado ningún requisito funcional sin cubrir en el código. Lo pendiente es **validar**: ejecutar las comprobaciones y confirmar que pasan. Hay históricos de ejecuciones previas con fallos en `Test_TasksCreate_*` (por CSRF en los tests de escritura); el commit `d222c26` añade el token CSRF a esos tests, pero no consta una ejecución posterior en verde.

**Propuesta:** no rediseñar. Ejecutar las validaciones; corregir solo lo que falle dentro del alcance de la feature; después abrir el PR con los commits existentes.

## Alcance

**Incluido:**

- Verificar que los criterios de aceptación están cubiertos por código y tests de la rama.
- Ejecutar `gofmt -l .`, `go vet ./...`, `go test -p 1 ./...` y `go build ./...` con `GO_ENV=test`.
- PR con los archivos de la feature y de este plan.

**Excluido (no entra en el PR):**

- `config/adw.yml`: modificado en el árbol de trabajo; es configuración local del trigger.
- `.txt`, `.x`, `migrations/schema.sql` (vacío), `.claude/skills/ship/`: no son parte de la feature.
- Los demás planes sin confirmar en `plans/` (más de 100 archivos sin trackear sobre este mismo issue). Solo entra este plan.

## Diseño por capas

- **Persistencia:** migración `add_column` nullable de tipo `date`. Sin índice: no hay consultas por fecha.
- **Modelo:** `*time.Time` para representar ausencia de fecha. Se guarda como día (medianoche UTC del `time.Parse`).
- **Acción y ruta:** `POST /tasks` y `PUT /tasks/{task_id}` leen `due_date`. Rutas existentes, sin rutas nuevas.
- **Renderizado:** Plush muestra el día con `Format`; el formulario de edición hace prefill con `YYYY-MM-DD`.
- **CSS:** solo ajuste visual de `.task-due` y del input de fecha.
- **JavaScript:** no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` | Columna `due_date` |
| `models/task.go` | Campo `DueDate` |
| `actions/app.go` | Rutas (sin cambios) |
| `actions/tasks.go` | `parseDueDate`, alta y edición con fecha |
| `templates/tasks/index.plush.html` | Alta con fecha y visualización |
| `templates/tasks/edit.plush.html` | Edición con fecha |
| `public/assets/app.css` | Estilos de fecha |
| `actions/tasks_test.go`, `models/task_test.go` | Tests de comportamiento |
| `database.yml` | BD de test `todo_test` |

## Implementación paso a paso

1. Confirmar que la rama está limpia en lo relativo a la feature (`git status`), dejando fuera los cambios ajenos listados en Alcance.
2. Ejecutar la validación (sección siguiente). Si todo pasa, no hay cambios de código.
3. Si algún test de la feature falla, corregir el mínimo en el archivo implicado, sin cambiar el comportamiento acordado en los criterios.
4. Hacer stage explícito de: migraciones de due_date, `models/task.go`, `actions/tasks.go`, `actions/tasks_test.go`, `models/task_test.go`, plantillas de tasks, `public/assets/app.css` y este plan. No incluir `config/adw.yml`.
5. Commits solo si hay cambios nuevos; mensajes Conventional Commits.

## Tests y validación

**Condiciones previas:**

- PostgreSQL escuchando en `127.0.0.1:5432` (el puerto está abierto en la sondeo de esta fase).
- Base `todo_test` con migraciones aplicadas, incluida `20261022120000`.
- Variable `GO_ENV=test` (los tests escriben en `development` si no está definida; ver memoria `go-tests-need-go-env-test`).

**Comandos:**

```
gofmt -l .
go vet ./...
GO_ENV=test go test -p 1 ./...
go build ./...
```

**Pruebas que demuestran el comportamiento:**

- Alta: `Test_TasksCreate_Valid`, `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksCreate_EmptyTitle_Invalid`, `Test_TasksCreate_OnlySpaces_Invalid`.
- Edición: `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksEdit_PrefillsDueDate`.
- Listado y toggle: `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate`.
- Modelo: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.

Si la rama ya tiene los tests, no se añaden tests nuevos. Si un criterio de la lista no tiene test, se añade uno siguiendo el patrón de `ActionSuite` y `ModelSuite`, sin ampliar el alcance.

## Riesgos y recuperación

- **Tests en rojo por CSRF:** los fallos históricos de `Test_TasksCreate_*` apuntan a esto. Si reaparecen, revisar que el formulario de test envía el token (commit `d222c26`).
- **Zona horaria:** `time.Parse` sin zona devuelve UTC; `Format` en Plush lo muestra tal cual. Si la BD devuelve la fecha con otra zona, el día podría desplazarse. Verificar con `Test_Task_DueDate_DayOnly_Persists`.
- **Tests contra `development`:** sin `GO_ENV=test` escriben datos reales locales. Usar siempre la variable.
- **Recuperación:** si la validación falla por algo fuera de la feature, detener y reportar sin tocar el alcance. La migración tiene `down.fizz` para revertir la columna en la BD local.

## Estado de la ejecución

Validado el 2026-10-24 con `GO_ENV=test` sobre `feat/issue-3-add-task-due-date`:

- `go vet ./...`: sin errores. `go build ./...`: correcto.
- `go test -p 1 ./actions/ ./models/` (sin caché): `Test_ActionSuite` y `Test_ModelSuite` en verde, incluidos los 36 casos de fecha límite listados arriba.
- `gofmt -l .`: marca archivos del árbol de trabajo por finales CRLF (`core.autocrlf=true`). Los blobs commiteados de la feature están limpios.
- `go test ./adws/`: falla (`Could not resolve to a Repository with the name 'example/todo'`). Falla igual en `main`, no está en el alcance de la feature y no se tocó.

El PR abierto para esta rama es el #5 y no se duplica.
