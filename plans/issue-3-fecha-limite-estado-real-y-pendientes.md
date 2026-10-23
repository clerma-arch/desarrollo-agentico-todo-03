# Feature: fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título `añadir fechaimie alas task`, cuerpo `vale un dia sin hora`. El título es ambiguo; la lectura adoptada es **añadir a cada tarea una fecha límite opcional con precisión de día y sin hora**.

Requisitos explícitos (del cuerpo de la issue):

1. Una tarea puede tener una fecha límite, solo día, sin hora.

Inferencias (no pedidas de forma explícita):

2. Al crear una tarea se puede indicar la fecha; si queda vacía, la tarea no tiene fecha.
3. Al editar se puede ver, cambiar, añadir o borrar la fecha.
4. La lista muestra la fecha de las tareas que la tienen.
5. Una fecha no válida se rechaza con mensaje de error y no se guarda.
6. Completar o reabrir una tarea conserva su fecha.

Fuera de la lectura: recordatorios, vencimientos visuales, ordenación o filtros por fecha, y hora de la tarea.

## Comportamiento actual y propuesta

Estado comprobado en esta sesión sobre `feat/issue-3-add-task-due-date` (HEAD `0f5ce4d`, base `main`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. Tiene `.down.fizz`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` solo valida el título.
- `actions/tasks.go`: `parseDueDate` (vacío = nil, formato `2006-01-02`). `TasksCreate` y `TasksUpdate` leen `due_date` en una struct de entrada, añaden el error `due_date` si no parsea y persisten con `tx.Create` / `tx.Update`. `TasksToggle` no toca el campo.
- `templates/tasks/index.plush.html`: `input type="date"` en el alta, `Fecha límite: dd/mm/yyyy` (`.task-due`) y error `due_date`.
- `templates/tasks/edit.plush.html`: campo `Fecha límite` con etiqueta, valor precargado, ayuda y error.
- `public/assets/app.css`: clase `.task-due` (línea 183) y ajustes del input de fecha.

Verificado en esta sesión:

- `go vet ./...`: OK.
- `go build ./...`: OK.
- `GO_ENV=test go test -count=1 -p 1 ./models/ ./actions/`: `ok todo/models`, `ok todo/actions`.

Cobertura por criterio (nombres de prueba existentes):

| Criterio | Pruebas |
|----------|---------|
| 1 | `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists` |
| 2 | `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_Valid` |
| 3 | `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate` |
| 4 | `Test_TasksIndex_ShowsDueDate` |
| 5 | `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate` |
| 5b (fallo no pierde datos) | `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate` |
| 6 | `Test_TasksToggle_KeepsDueDate` |

Propuesta: no hay cambios funcionales. Se verifica que la rama cumple los criterios y se deja lista para PR. Solo se toca código si una verificación falla de verdad, con el cambio mínimo y su prueba.

## Alcance

Incluido:

- Confirmar que los criterios 1 a 6 se cumplen con las pruebas existentes.
- Ejecutar las validaciones pertinentes con `GO_ENV=test`.
- Revisar que el PR no incluya cambios ajenos.

Excluido:

- Rediseñar la implementación presente.
- Recordatorios, vencimientos, ordenación, filtros, hora de la tarea.
- Los fallos previos de `adws/trigger_test.go`: dependen de un `gh` real y fallan también en `main`.
- Cambios sin confirmar ajenos a la feature: `config/adw.yml` (modificado), `.claude/skills/ship/`, `.txt`, `migrations/schema.sql` y los `plans/` sin seguir. No entran en el PR.

## Diseño por capas

- **Persistencia**: columna `due_date date null` en `tasks`. Migración ya presente; no requiere cambios.
- **Modelo**: `Task.DueDate *time.Time`; nil = sin fecha. Sin cambios.
- **Acción y ruta**: `POST /tasks` y `PUT /tasks/{task_id}` leen `due_date`. Sin cambios de rutas.
- **Renderizado**: Plush en `index` y `edit`. Sin cambios.
- **CSS**: solo `.task-due` y el input de fecha. Sin cambios.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|------|-----------------|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y quita `due_date`. |
| `models/task.go` | Campo `DueDate`. |
| `models/task_test.go` | Pruebas de modelo de la fecha. |
| `actions/tasks.go` | `parseDueDate`, alta, edición, toggle. |
| `actions/tasks_test.go` | Pruebas de alta, edición, validación, listado y toggle. |
| `templates/tasks/index.plush.html` | Alta y visualización. |
| `templates/tasks/edit.plush.html` | Campo de edición. |
| `public/assets/app.css` | Estilos de la fecha. |

## Implementación paso a paso

1. Confirmar con `git status --short` que solo quedan fuera del PR los cambios ajenos listados en Alcance.
2. Ejecutar las validaciones de la sección siguiente.
3. Si todas pasan y los criterios 1 a 6 están cubiertos, no modificar código; pasar a PR.
4. Si una verificación falla, añadir la prueba que lo reproduce y el cambio mínimo en la capa afectada. No ampliar el alcance. Si el fallo obliga a un cambio de diseño, detenerse y reportarlo.

## Tests y validación

Condiciones:

- PostgreSQL accesible en `127.0.0.1` (o `DB_HOST`) con la base `todo_test` y migraciones de test aplicadas (`soda migrate -e test`), como en CI.
- `GO_ENV=test`. Sin esa variable, las pruebas de acciones devuelven 403 por CSRF. Ocurre también en `main`.

Comandos:

- `gofmt -l .`: en Windows marca archivos con CRLF en el worktree. Para verificar el contenido real, comparar contra el índice (`git ls-files --eol`) o ejecutar sobre una copia sin `\r`.
- `go vet ./...`
- `GO_ENV=test go test -count=1 -p 1 ./models/ ./actions/`
- `go build ./...`
- `GO_ENV=test go test -count=1 -p 1 ./...`: el paquete `adws` falla en `main` y en la rama porque invoca `gh` real; se reporta como fallo previo, no como regresión.

## Riesgos y recuperación

- **Zona horaria**: `time.Parse` produce medianoche UTC. Si el driver devolviera otro huso, el día mostrado podría cambiar. Las pruebas de persistencia lo cubren; si falla, fijar el formateo a UTC en modelo o plantilla.
- **Navegadores sin input de fecha**: el servidor valida el formato igualmente.
- **Planes duplicados**: `plans/` contiene decenas de planes previos de esta issue. Este plan no los sustituye; registra el estado comprobado y lo pendiente.
- **Recuperación**: la migración tiene `down`; `soda migrate down` en el entorno correspondiente.
