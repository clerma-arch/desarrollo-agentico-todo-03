# Feature: fecha límite opcional (solo día) en las tareas (issue #3) — estado actual y cierre

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

Estado leído en el código de `feat/issue-3-add-task-due-date` (HEAD `1fc855a`). **La funcionalidad ya está implementada en la rama.** Este plan no propone reimplementarla; registra qué cubre cada criterio y qué queda por validar antes del PR.

Cobertura por criterio (código y test asociado):

- Criterio 1 y 2: `TasksCreate` en `actions/tasks.go`; `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_Valid`.
- Criterio 3: `parseDueDate` (`time.Parse("2006-01-02")`) en `actions/tasks.go`; `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate` (muestra `22/10/2026`); `Test_Task_DueDate_DayOnly_Persists` en `models/task_test.go`.
- Criterio 4: `TasksCreate` y `TasksUpdate` devuelven `dueDateInput` (texto crudo) en error. Lo añade el commit `1fc855a`, que cierra el hueco que señalaba `plans/issue-3-add-task-due-date-cerrar-criterios.md`. Tests: `Test_TasksCreate_InvalidDueDate` y `Test_TasksUpdate_InvalidDueDate`, ahora con `value="no-es-fecha"`.
- Criterio 5: `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- Criterio 6: `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`.
- Criterio 7: `Test_TasksToggle_KeepsDueDate`.

Piezas ya presentes:

- Persistencia: migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` (`add_column("tasks", "due_date", "date", {"null": true})`).
- Modelo: `models/task.go`, campo `DueDate *time.Time` con tags `db:"due_date"` y `form:"due_date"`.
- Plantillas: `templates/tasks/index.plush.html` (input de alta y la fecha bajo el título en la lista) y `templates/tasks/edit.plush.html` (campo de edición con ayuda y error). Ambas usan `dueDateInput` si no está vacío y, si no, el formato de `task.DueDate`.
- CSS: `public/assets/app.css`, clases `.task-due` y `.new-task-row input[type="date"]`.

Propuesta: ninguna modificación de código. Solo validar que la rama cumple los criterios y dejarla lista para el PR.

## Alcance

Incluido:

- Verificar que los tests de `actions/` y `models/` pasan con la rama actual.
- Verificar que `go vet`, `go build` y `gofmt` sobre los archivos Go modificados no dan problemas.
- Dejar el PR con los commits de este trabajo.

Excluido:

- Reimplementar la fecha límite, la migración, el modelo o la lista.
- Hora, zona horaria, recordatorios, filtros u orden por fecha.
- Archivos sin confirmar que no son de este cambio: `config/adw.yml` (modificado), `.claude/skills/ship/`, `.txt`, `migrations/schema.sql`, y los `plans/` untracked distintos de los de este plan. No entran en el PR.
- Fallos de `adws/` que ya existen en `main`.
- Conversión de finales de línea CRLF: `gofmt -l .` marca casi todos los archivos por el entorno Windows, no por este cambio.

## Diseño por capas

- **Persistencia**: sin cambios. Columna `due_date` `date` nullable ya migrada.
- **Modelo**: sin cambios. La validación de formato vive en la acción (`parseDueDate`), no en el modelo.
- **Acción y ruta**: sin cambios pendientes. `TasksIndex`, `TasksCreate`, `TasksEdit`, `TasksUpdate` definen `dueDateInput`; `TasksToggle` y `TasksDestroy` no cambian.
- **Renderizado**: sin cambios pendientes en `index.plush.html` ni `edit.plush.html`.
- **CSS**: sin cambios pendientes.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `actions/tasks.go` | `parseDueDate`, `dueDateInput` en los cuatro handlers de render |
| `actions/tasks_test.go` | Tests de alta, edición, error y conservación de la fecha |
| `models/task.go`, `models/task_test.go` | Campo `DueDate` y tests de persistencia (día sin hora, opcional) |
| `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` | Columna `due_date` |
| `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html` | Input `due_date` y visualización |
| `public/assets/app.css` | Estilos de la fecha |

## Implementación paso a paso

1. Comprobar que la rama está en `1fc855a` (o posterior) y que el working tree solo tiene los archivos sin confirmar listados en Alcance.
2. Ejecutar los tests de `actions/` y `models/` (ver Tests y validación).
3. Si todo pasa, no hay cambios de código: pasar a `/ship` para el PR. Si algún test falla, parar e informar con la salida; no implementar un arreglo sin volver a planificar.

## Tests y validación

Condiciones necesarias:

- PostgreSQL local escuchando en `127.0.0.1:5432`, base `todo_test`, usuario `postgres`/`postgres`, con migraciones aplicadas (`soda migrate -e test`).
- `GO_ENV=test` en el entorno. Sin esta variable, `actions/` falla con 403 por CSRF; CI la define en `.github/workflows/ci.yaml`.

Comprobaciones:

- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/` debe pasar.
- `go vet ./...` y `go build ./...` deben pasar.
- `gofmt -l actions/tasks.go actions/tasks_test.go models/task.go models/task_test.go` debe quedar vacío, o explicar cualquier salida que sea solo CRLF. No usar `gofmt -l .` como criterio.
- Fallos de `adws/` ya existentes en `main`: no son regresión de esta feature.
- Migración up/down: opcional. Si el entorno lo permite, `soda migrate -e test`, `soda migrate down -e test` y volver a aplicarla.

Tests que demuestran el comportamiento: los listados en "Comportamiento actual y propuesta". No se añade ninguno nuevo.

## Riesgos y recuperación

- **Estado no verificado en ejecución**: esta sesión no ha corrido la suite. "Cumple" es lectura del código y de los tests, no una ejecución. Queda confirmado solo tras el paso 2.
- **Variable indefinida en Plush**: si una ruta de render no define `dueDateInput`, el template falla. Las cuatro rutas la definen; los tests de `Test_TasksIndex_*` y `Test_TasksEdit_*` lo cubren.
- **Escapado HTML**: `<%= %>` escapa automáticamente; no se ha verificado un valor malicioso concreto en el atributo `value`.
- **Zona horaria**: `time.Parse` sin zona produce medianoche UTC y el formateo no cambia el día. Lo cubre `Test_Task_DueDate_DayOnly_Persists`.
- **Rollback**: revertir la rama a `main` quita la funcionalidad; la migración `down` elimina la columna. No hay pérdida de datos fuera de la columna nueva.
- **Incertidumbre 1**: no se ha confirmado si `1fc855a` pasa la suite completa.
- **Incertidumbre 2**: `config/adw.yml` modificado en el working tree no se ha revisado; se asume que no es parte de este cambio.
