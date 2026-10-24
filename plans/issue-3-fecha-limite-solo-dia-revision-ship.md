# Feature: fecha límite opcional (solo día) en las tareas (issue #3) — revisión para PR

## Encargo y criterios de aceptación

Encargo de la issue #3 (`gh issue view 3`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"):

- Explícito: las tareas pueden tener una fecha límite.
- Explícito: la fecha es solo día, sin hora.
- Inferido: la fecha es opcional; una tarea sin fecha sigue siendo válida.
- Inferido: la fecha se crea, ve, edita y borra desde la interfaz existente (lista y edición de tareas).

Criterios de aceptación:

1. Al crear una tarea, el formulario de la lista acepta una fecha opcional (`due_date`, tipo `date`).
2. Una tarea creada sin fecha se guarda con `due_date` nulo.
3. Una fecha válida `YYYY-MM-DD` se guarda como día sin hora y se muestra como `dd/mm/aaaa`.
4. Una fecha no válida no crea la tarea, muestra el error en `due_date` y conserva en el formulario el texto que escribió la persona.
5. Un título vacío o demasiado largo se sigue rechazando, y la fecha introducida se conserva.
6. Editar permite cambiar o borrar la fecha; un error de validación conserva el título y la fecha introducidos.
7. Completar o reabrir una tarea no modifica su fecha límite.

## Comportamiento actual y propuesta

**La funcionalidad ya está implementada en `feat/issue-3-add-task-due-date`** (HEAD `00c9983`, con `1fc855a` como último cambio de código). Esta revisión no propone código nuevo: comprueba que la implementación cumple los criterios y deja el branch listo para el PR contra `main`.

Verificado en esta revisión (lectura de código):

- Los nombres de test que cubren cada criterio existen en `actions/tasks_test.go` y `models/task_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate`, `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
- Persistencia: `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` añade `due_date` `date` nullable.
- Modelo: `models/task.go`, `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- Acción: `parseDueDate` (`time.Parse("2006-01-02")`) en `actions/tasks.go`; `TasksCreate` y `TasksUpdate` pasan `dueDateInput` al re-renderizar con error; `TasksIndex` y `TasksEdit` lo definen vacío.
- Plantillas: `templates/tasks/index.plush.html` (input de alta y `Fecha límite: dd/mm/aaaa` bajo el título) y `templates/tasks/edit.plush.html` (campo, ayuda y error).
- CSS: `public/assets/app.css`, clases `.task-due` y `.new-task-row input[type="date"]`.
- Entorno: PostgreSQL responde en `127.0.0.1:5432`.

**No verificado todavía**: que la suite pase en esta sesión. Ver "Tests y validación".

Propuesta: ningún cambio de código ni de tests. Solo ejecutar la validación y, si pasa, abrir el PR.

## Alcance

Incluido:

- Ejecutar los tests de `actions/` y `models/`, `go vet`, `go build` y `gofmt` sobre los archivos de la feature.
- Abrir el PR de `feat/issue-3-add-task-due-date` contra `main`, con los commits que ya tiene la rama.

Excluido:

- Reimplementar la fecha límite, la migración, el modelo o la lista.
- Hora, zona horaria, recordatorios, filtros u orden por fecha.
- Cambios sin confirmar que no son de esta feature: `config/adw.yml` (modificado), `.txt`, `migrations/schema.sql`, `.claude/skills/ship/`, y los `plans/` untracked distintos de los de esta issue. No entran en el PR ni en el stage.
- Fallos de `adws/` que ya existen en `main`, si aparecen en `go test ./...`.
- Conversión de finales de línea. Ver Riesgos.

## Diseño por capas

- **Persistencia**: sin cambios. Columna `due_date` `date` nullable ya migrada.
- **Modelo**: sin cambios. La validación de formato vive en la acción (`parseDueDate`), no en el modelo.
- **Acción y ruta**: sin cambios. Rutas `POST /tasks`, `GET /tasks/{id}/edit`, `PUT /tasks/{id}` ya existen en `actions/app.go`.
- **Renderizado**: sin cambios.
- **CSS**: sin cambios.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `actions/tasks.go` | `parseDueDate`, `dueDateInput` en los handlers de render, error de `due_date` |
| `actions/tasks_test.go` | Tests de alta, edición, error y conservación de la fecha |
| `models/task.go`, `models/task_test.go` | Campo `DueDate` y tests de persistencia |
| `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` | Columna `due_date` |
| `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html` | Input `due_date` y visualización |
| `public/assets/app.css` | Estilos de la fecha |

## Implementación paso a paso

1. Comprobar que la rama está en `00c9983` o posterior y que el working tree solo tiene los cambios sin confirmar listados en Alcance.
2. Ejecutar la validación de "Tests y validación".
3. Si todo pasa, no hay cambios de código: continuar con el PR (`/ship`), haciendo stage solo de los archivos de la feature y su plan.
4. Si algún test falla, parar e informar con la salida. No arreglar sin volver a planificar.

## Tests y validación

Condiciones necesarias:

- PostgreSQL en `127.0.0.1:5432`, base `todo_test`, usuario `postgres`/`postgres`, con migraciones aplicadas (`soda migrate -e test`). PostgreSQL responde; falta confirmar que la base y las migraciones existen.
- `GO_ENV=test` en el entorno. Sin ella, `actions/` falla con 403 por CSRF, según el plan previo (`plans/issue-3-add-task-due-date-estado-actual.md`). Confirmar en `.github/workflows/ci.yaml` antes de depender de ello.

Comprobaciones:

- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/` debe pasar.
- `go vet ./...` y `go build ./...` deben pasar.
- `gofmt`: `gofmt -l .` marca casi todos los archivos en este entorno Windows. Usar en su lugar la comprobación de los archivos de la feature, ignorando CRLF: `tr -d '\r' < <archivo> | gofmt -l` debe quedar vacío para `actions/tasks.go`, `actions/tasks_test.go`, `models/task.go`, `models/task_test.go`.
- Fallos de `adws/` que ya existan en `main`: no son regresión de esta feature.
- Migración up/down: opcional. Si el entorno lo permite, `soda migrate down -e test` y volver a aplicarla.

Tests que demuestran el comportamiento: los listados en "Comportamiento actual y propuesta". No se añade ninguno nuevo.

## Riesgos y recuperación

- **Estado no ejecutado**: esta revisión lee código y tests, no los ejecuta. "Cumple" queda confirmado solo tras el paso 2.
- **Variable no definida en Plush**: si una ruta de render no define `dueDateInput`, el template falla. Las cuatro rutas la definen; los tests de `Test_TasksIndex_*` y `Test_TasksEdit_*` lo cubren.
- **Escapado HTML**: `<%= %>` escapa automáticamente; no se ha probado un valor malicioso concreto en el atributo `value`.
- **Zona horaria**: `time.Parse` sin zona produce medianoche UTC; el formateo no cambia el día. Lo cubre `Test_Task_DueDate_DayOnly_Persists`.
- **Finales de línea**: `gofmt` marca los archivos por CRLF del entorno Windows. Antes de tomar la salida como error, comparar con `main`; si `main` también la produce, no es de esta feature.
- **Rollback**: revertir la rama deja `main` como estaba; la migración `down` elimina la columna, sin pérdida de datos fuera de ella.
- **Incertidumbre 1**: no se ha confirmado que la suite completa pase en esta sesión.
- **Incertidumbre 2**: `config/adw.yml` modificado no se ha revisado; se asume que no es parte de la feature.
- **Incertidumbre 3**: la rama acumula commits de documentación de `plans/` además del código. Se incluyen en el PR, o se deciden fuera, antes de abrirlo.
