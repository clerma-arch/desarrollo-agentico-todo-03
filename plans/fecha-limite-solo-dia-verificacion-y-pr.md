# Feature: verificar y cerrar la fecha límite (solo día) de las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". Título ambiguo ("fechaimie"); la lectura adoptada es **añadir a cada tarea una fecha límite opcional, con precisión de día y sin hora**.

Requisitos explícitos (cuerpo de la issue):
- La fecha es solo de día (`YYYY-MM-DD`), sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional: una tarea puede no tener fecha.
- La fecha se puede indicar al crear y al editar una tarea.
- La fecha se muestra en la lista (`DD/MM/YYYY`) y se precarga en el formulario de edición.

Criterios de aceptación:
1. Crear una tarea sin fecha funciona como antes.
2. Crear una tarea con fecha `2026-10-22` la guarda como día y la lista la muestra como `22/10/2026`.
3. Editar una tarea permite cambiar o borrar la fecha sin alterar el título ni `Completed`.
4. Una fecha mal formada no se guarda: la respuesta es 422 con mensaje en `due_date`, nunca 500.
5. Alternar completada/pendiente conserva la fecha.
6. La migración añade la columna nullable; las tareas existentes quedan sin fecha.

## Comportamiento actual y propuesta

**Estado: la funcionalidad ya está implementada en la rama `feat/issue-3-add-task-due-date`** (22 commits sobre `main`, rama al día con `origin/feat/issue-3-add-task-due-date`). No hay código de aplicación pendiente conocido. Este plan se centra en verificar la implementación, comprobar que cumple los criterios y dejar el PR listo.

Implementado en la rama:
- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`; el `.down.fizz` hace `drop_column`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` no cambia.
- `actions/tasks.go`:
  - `parseDueDate(raw string) (*time.Time, error)`: cadena vacía → `nil`; si no, `time.Parse("2006-01-02")`.
  - `TasksCreate` y `TasksUpdate` leen `title` y `due_date` con un struct de entrada, validan con `task.Validate(tx)` y añaden el error `due_date` si el parseo falla. En error reenvían `dueDateInput` para no perder lo escrito.
  - `TasksIndex` y `TasksEdit` fijan `dueDateInput` vacío.
- `templates/tasks/index.plush.html`: `<input type="date" name="due_date">` en el formulario de alta, muestra `Fecha límite: DD/MM/YYYY` junto al título y el error `due_date`.
- `templates/tasks/edit.plush.html`: campo `Fecha límite` con ayuda "Opcional. Solo indica el día, sin hora." y error `due_date`.
- `public/assets/app.css`: 11 líneas para las clases de la fecha y de la ayuda de campo.
- Tests: `actions/tasks_test.go` (suite `ActionSuite`) y `models/task_test.go` (suite `ModelSuite`).

Tests de fecha ya presentes (no ejecutados en este plan):
- `actions/tasks_test.go`: `Test_TasksToggle_KeepsDueDate`, `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`.
- `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.

Comportamiento que hay que confirmar leyendo o ejecutando (no asumir):
- El día guardado no cambia de fecha por zona horaria (se parsea en UTC con `time.Parse` y se formatea con `Format`).
- Con fecha inválida en el alta, el título escrito y la fecha inválida se conservan en el formulario de nuevo.

## Alcance

Incluido:
- Verificar la implementación existente contra los criterios de aceptación.
- Ejecutar las validaciones habituales y, si falta algún caso del criterio 1-6 sin test, añadir **solo** ese test siguiendo el patrón de `ActionSuite`/`ModelSuite`.
- Preparar el PR contra `main` con los archivos del cambio y el plan.

Excluido:
- Rediseñar el campo (no hay hora, zona horaria, recordatorios, filtros por fecha ni ordenación por fecha).
- Cambiar `config/adw.yml`, `.claude/skills/ship/`, `.txt`, `migrations/schema.sql` ni los planes sin seguimiento de otras tareas (ver "Archivos relevantes").
- Reescribir la historia de commits de la rama.

## Diseño por capas

- **Persistencia**: columna `due_date date NULL` en `tasks`. Migración ya creada con timestamp posterior a la última (`20260901000001_seed_tasks`).
- **Modelo**: campo `*time.Time` (nil = sin fecha). Usar `*time.Time` evita escribir la fecha cero `0001-01-01` cuando no hay fecha.
- **Acción y ruta**: `POST /tasks` y `PUT /tasks/{task_id}` ya existen; sin rutas nuevas. Entrada por formulario (`due_date`).
- **Render**: `index.plush.html` y `edit.plush.html`. Sin JavaScript: el `input type="date"` nativo basta.
- **CSS**: solo clases de presentación para la fecha y la ayuda del campo.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date`. |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna. |
| `models/task.go` | Campo `DueDate *time.Time`. |
| `models/task_test.go` | Tests de modelo de la fecha opcional y de solo día. |
| `actions/tasks.go` | `parseDueDate`, alta y edición con `due_date`, reenvío de la entrada en errores. |
| `actions/tasks_test.go` | Tests de acción: alta, edición, errores, alternar y render. |
| `templates/tasks/index.plush.html` | Campo de alta y visualización de la fecha en la lista. |
| `templates/tasks/edit.plush.html` | Campo de edición y ayuda. |
| `public/assets/app.css` | Estilos de la fecha y de la ayuda. |
| `actions/app.go` | Rutas `/`, `/tasks`, `/tasks/{task_id}` (sin cambios). |
| `database.yml` | Base `todo_test` en PostgreSQL, `DB_HOST` o `127.0.0.1`. |

Archivos del branch que **no** forman parte del PR:
- `config/adw.yml` (modificado sin confirmar): cambia `repo` y el `prompt` del trigger de issues. Es una configuración local ajena a la funcionalidad.
- `.claude/skills/ship/`, `.txt`, `migrations/schema.sql` (sin seguimiento): no son del cambio.
- Planes sin seguimiento en `plans/` distintos del actual: otros borradores.

## Implementación paso a paso

No hay código nuevo que escribir por defecto. Pasos:

1. Confirmar el estado: `git status --short --branch` debe mostrar solo los archivos excluidos de la lista anterior como no staged.
2. Preparar PostgreSQL con la base `todo_test` y las migraciones aplicadas (ver "Tests y validación").
3. Ejecutar las validaciones y registrar la salida.
4. Leer los tests de fecha contra los criterios 1-6. Si un criterio no tiene test, añadir ese test (y solo ese) en el archivo que corresponda, con el patrón de la suite existente.
5. Si algún test falla, detenerse y reportar la salida; no cambiar el alcance.
6. Stage explícito (nunca `git add .` ni `git add -A`) de:
   - `migrations/20261022120000_add_due_date_to_tasks.up.fizz`
   - `migrations/20261022120000_add_due_date_to_tasks.down.fizz`
   - `models/task.go`, `models/task_test.go`
   - `actions/tasks.go`, `actions/tasks_test.go`
   - `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`
   - `public/assets/app.css`
   - `plans/fecha-limite-solo-dia-verificacion-y-pr.md`
7. Si hay cambios de los pasos 4 o 6 aún sin commit, crear un commit Conventional Commits solo para ellos (mensaje en español, con referencia a `issue #3`). No crear commits ajenos a esta rama de la issue.
8. `git push origin feat/issue-3-add-task-due-date` (sin `--force`).
9. `gh pr create --base main` con título Conventional Commits (`feat(tasks): añadir fecha límite solo día a las tareas (issue #3)`), descripción con resumen, ruta de este plan, tests añadidos y resultado de la validación, y la línea final de atribución indicada por el flujo `/ship`.

## Tests y validación

Requisitos previos:
- PostgreSQL accesible desde `DB_HOST` (por defecto `127.0.0.1`) con usuario `postgres`/`postgres` y base `todo_test`.
- Migraciones aplicadas sobre `todo_test` (incluida `add_due_date_to_tasks`). Si `buffalo test` usa `migrations/schema.sql` en vez de las migraciones, ese archivo sin seguimiento está desactualizado (no tiene `due_date`) y debe regenerarse solo para la ejecución local; **no** añadirlo al PR. Esto es una incertidumbre a confirmar al ejecutar.

Comandos y criterio de paso:
- `gofmt -l .`: sin salida.
- `go vet ./...`: sin errores.
- `go build ./...`: compila.
- `go test -p 1 ./...`: todos los tests pasan, incluidos los de fecha listados arriba.

Cobertura esperada por criterio:
- Criterio 1: `Test_TasksCreate_Valid` (existente, sin fecha) y `Test_Task_DueDate_Optional`.
- Criterio 2: `Test_TasksCreate_WithDueDate`, `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3: `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksUpdate_PersistsTitleOnly` (existente).
- Criterio 4: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5: `Test_TasksToggle_KeepsDueDate`.
- Criterio 6: revisión de la migración (`up`/`down`) en lectura; no hay test automático de migraciones en el repositorio.
- Errores que conservan la entrada: `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.

Verificación manual opcional: `buffalo dev` (o `/env:start`), crear una tarea con fecha, editarla, borrar la fecha y comprobar la lista y los mensajes de error.

## Riesgos y recuperación

- **Validación de título fuera del flujo de validación de Pop**: `TasksCreate` y `TasksUpdate` pasan de `tx.ValidateAndCreate` / `tx.ValidateAndUpdate` a `task.Validate(tx)` seguido de `tx.Create` / `tx.Update`. Los hooks de validación de creación/actualización de Pop no se ejecutan de la misma forma. Confirmar que el modelo no depende de ellos (`Validate` solo valida título). Si se detecta lo contrario, reportarlo y no ampliar el alcance.
- **Desfase de zona horaria**: `time.Parse` produce `00:00 UTC`; PostgreSQL guarda `date`, así que el día no debería cambiar. Cubrirlo con la prueba de solo día existente; si falla, detenerse.
- **Migración con datos**: la columna es nullable y sin default; las filas existentes quedan `NULL`. Revertir con el `.down.fizz` si hace falta.
- **Fichero `migrations/schema.sql` sin seguimiento**: si los tests lo usan, puede dar fallos engañosos. No se commitea.
- **Planes duplicados en `plans/`**: el repositorio ya tiene varios planes de esta issue (algunos sin seguimiento). Solo se incluye este plan en el PR; los demás no se tocan.
- **Recuperación**: si un test falla tras el push, no forzar el push; corregir con un commit nuevo o cerrar el PR.

Incertidumbres pendientes:
- Si PostgreSQL y `todo_test` están disponibles en este equipo (no comprobado al escribir el plan).
- Si `buffalo test` carga el esquema desde las migraciones o desde `migrations/schema.sql`.
- Si los hooks de Pop eran relevantes para el modelo `Task`.
