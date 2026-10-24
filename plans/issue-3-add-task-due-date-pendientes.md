# Feature: fecha límite opcional (solo día) en las tareas (issue #3)

## Encargo y criterios de aceptación

Encargo de la issue #3 (título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"):

- Explícito: las tareas pueden tener una fecha límite.
- Explícito: la fecha es solo día, sin hora.
- Inferido: la fecha es opcional; una tarea sin fecha sigue siendo válida.
- Inferido: la fecha se puede crear, ver, editar y borrar desde la interfaz existente.

Criterios de aceptación:

1. Al crear una tarea, el formulario de la lista acepta una fecha opcional (`due_date`, tipo `date`).
2. Una tarea creada sin fecha se guarda con `due_date` nulo.
3. Una fecha válida `YYYY-MM-DD` se guarda como día sin hora y se muestra como `dd/mm/aaaa`.
4. Una fecha no válida no crea la tarea y muestra un error en `due_date`; el título y la fecha introducidos se conservan.
5. Un título vacío o demasiado largo se sigue rechazando, y la fecha introducida se conserva en el formulario.
6. Editar una tarea permite cambiar o borrar la fecha; un error de validación conserva la fecha introducida.
7. Completar o reabrir una tarea no modifica su fecha límite.

## Comportamiento actual y propuesta

Estado verificado en la rama `feat/issue-3-add-task-due-date` (commits `d13ec2c` … `786fcc4`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: columna `due_date` `date` nullable.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en día sin hora; `TasksCreate` y `TasksUpdate` validan con `task.Validate(tx)` y añaden el error `due_date` a `verrs`; se persiste con `tx.Create` / `tx.Update`.
- `templates/tasks/index.plush.html`: input `date` en el formulario de alta, texto `Fecha límite: dd/mm/aaaa` en cada tarea, error `due_date`.
- `templates/tasks/edit.plush.html`: input `date` prellenado, texto de ayuda y error `due_date`.
- `public/assets/app.css`: clases `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Tests añadidos en `actions/tasks_test.go` y `models/task_test.go` (ver sección de tests).

Con `GO_ENV=test`, `go test -count=1 -p 1 ./actions/ ./models/` pasa: `actions` ok (33 subpruebas PASS, 0 FAIL) y `models` ok.

Propuesta: no reimplementar. Este plan solo cierra lo que quede por verificar frente a los criterios de aceptación.

## Alcance

Incluido:

- Verificar cada criterio de aceptación contra el código y los tests de la rama.
- Corregir únicamente un fallo concreto que un criterio no cubra, si aparece.
- Comprobar que la migración sube y baja limpiamente en PostgreSQL de test.

Excluido:

- Hora, zona horaria o recordatorios.
- Filtros o orden por fecha límite.
- Cambios en `adws/` o en la configuración `config/adw.yml`.
- Los fallos de `adws/` que ya existen en `main` (llamadas reales a `gh` con `example/todo`).
- El formato de finales de línea CRLF de los archivos: `gofmt -l .` marca casi todos los archivos por CRLF del entorno Windows y no por el cambio.

## Diseño por capas

- **Persistencia**: migración `add_column("tasks", "due_date", "date", {"null": true})`. Ya existe. Revertir con `drop_column`.
- **Modelo**: `DueDate *time.Time`. Ya existe. La validación de la fecha no está en `Task.Validate` porque la conversión ocurre en la acción; ese es el patrón elegido.
- **Acción y ruta**: `TasksCreate`, `TasksUpdate` y `TasksToggle` (sin cambios de firma). Ya existen.
- **Renderizado**: `index` y `edit` con el input `date`, la visualización y los errores. Ya existe.
- **CSS**: solo estilos del texto de fecha y del input en la fila de alta. Ya existe.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añadir y quitar la columna |
| `models/task.go` | Campo `DueDate` y `Validate` |
| `models/task_test.go` | Pruebas de modelo para fecha opcional y persistencia de día |
| `actions/tasks.go` | `parseDueDate`, alta, edición y errores |
| `actions/tasks_test.go` | Pruebas de acción (alta, edición, errores, toggle, render) |
| `templates/tasks/index.plush.html` | Formulario de alta, visualización y errores |
| `templates/tasks/edit.plush.html` | Formulario de edición |
| `public/assets/app.css` | Estilos de la fecha |

## Implementación paso a paso

1. Partir de la rama `feat/issue-3-add-task-due-date` tal como está.
2. Verificar cada criterio de aceptación con los tests existentes, leyendo el test asociado.
3. Solo si un criterio no tiene cobertura o falla: añadir el test mínimo que lo demuestre y corregir el código en el archivo que corresponda.
4. Revisar que la migración `up` y `down` se aplican en la base de test.
5. No tocar archivos fuera de la lista de archivos relevantes.

## Tests y validación

Condiciones necesarias:

- PostgreSQL local en `127.0.0.1` con la base `todo_test` y usuario `postgres`/`postgres`, con las migraciones aplicadas (`soda migrate -e test`).
- `GO_ENV=test` en el entorno. Sin esta variable, `actions/` falla con 403 por CSRF; CI la define en `.github/workflows/ci.yaml`.

Comprobaciones:

- `go test -count=1 -p 1 ./actions/ ./models/` debe pasar.
- `go vet ./...` y `go build ./...` deben pasar.
- `gofmt -l` sobre los archivos modificados debe quedar vacío. No usar `gofmt -l .` como criterio: marca archivos por CRLF del entorno y los worktrees de `.claude/worktrees/`.
- Los fallos de `adws/` existen también en `main` y no se consideran regresión de esta feature.

Tests que demuestran el comportamiento (ya presentes):

- `models`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
- `actions`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate`.

## Riesgos y recuperación

- Zona horaria: `time.Parse` sin zona produce medianoche UTC. Al mostrar con `Format` no hay desplazamiento de día porque no se convierte de zona. Si la base o el driver devuelven la fecha con otra zona, el día mostrado podría cambiar; verificarlo con el test `Test_Task_DueDate_DayOnly_Persists`.
- Rollback: `drop_column("tasks", "due_date")` revierte la migración; el código sin la columna dejaría de compilar contra la base, por lo que hay que revertir migración y código juntos.
- Incertidumbre: no se ha ejecutado la migración `down` en esta sesión.
