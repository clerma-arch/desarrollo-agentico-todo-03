# Feature: fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Issue #3 del repositorio `clerma-arch/desarrollo-agentico-todo-03`: título "añadir fechaimie alas task", cuerpo "vale un dia sin hora", label `adw`.

Interpretación: cada tarea puede tener una fecha límite opcional que guarda solo el día, sin hora.

Requisitos explícitos (del cuerpo de la issue):

- La fecha es solo día, sin hora.

Inferencias (no están en la issue; se derivan del patrón del repositorio y de los commits previos de la rama):

- La fecha es opcional: una tarea puede no tener fecha.
- Se puede fijar al crear y cambiar o quitar al editar.
- La fecha se muestra en la lista y se rellena al editar.
- Completar o reabrir una tarea no modifica su fecha.

Criterios de aceptación:

1. Crear una tarea sin fecha la deja sin fecha (`NULL` en BD).
2. Crear una tarea con `due_date=YYYY-MM-DD` persiste ese día a medianoche, sin hora significativa.
3. Una fecha con formato inválido no se guarda y muestra el error `La fecha no es válida.` en el campo `due_date`.
4. Un error de título conserva la fecha válida introducida en el formulario de alta.
5. Editar puede fijar, cambiar y quitar la fecha; el formulario de edición la precarga.
6. La lista muestra `Fecha límite: DD/MM/AAAA` solo en las tareas que tienen fecha.
7. Completar y reabrir una tarea conserva su fecha.

## Comportamiento actual y propuesta

Estado observado en la rama `feat/issue-3-add-task-due-date` (10 commits sobre `main`, 15 archivos cambiados en el diff contra `main`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `down`.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` (formato `2006-01-02`, cadena vacía = sin fecha); `TasksCreate` y `TasksUpdate` leen `due_date`, validan con `task.Validate(tx)` y añaden el error `due_date` si el formato falla; luego usan `tx.Create` / `tx.Update`.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de alta y `<span class="task-due">` en cada tarea con fecha.
- `templates/tasks/edit.plush.html`: input `type="date"` con precarga, texto de ayuda y error.
- `public/assets/app.css`: clase `.task-due` y estilo del date input en `.new-task-row`.
- Tests: `models/task_test.go` (2 tests de fecha) y `actions/tasks_test.go` (9 tests de fecha: crear con y sin error, error de título con fecha en alta y edición, editar fijar/quitar/error, listar, precarga, completar).

Propuesta: no hay cambios funcionales pendientes. La implementación cubre los criterios 1–7. El plan consiste en verificar que el estado de la rama cumple la issue, ejecutar la validación en el entorno correcto y, si la validación pasa, cerrar la rama sin cambios de código.

Huecos menores detectados (no bloquean el PR; no están en la issue):

- En un error de formato de fecha, la plantilla recibe `task.DueDate = nil`, así que el input vuelve vacío y el usuario no ve lo que escribió. El mensaje de error sí aparece. Tratarlo sería un cambio de UX fuera del alcance de la issue.
- No hay indicador visual de vencimiento (tarea vencida). No se pidió.

## Alcance

Incluido:

- Verificar migración, modelo, acción, plantillas, CSS y tests ya presentes en la rama contra los criterios de aceptación.
- Ejecutar las validaciones con el entorno de test correcto (ver "Tests y validación").

Excluido:

- Cambiar el comportamiento de la fecha (zona horaria, vencimiento, ordenación por fecha, filtros).
- Corregir la UX de re-render tras error de fecha (hueco menor de arriba).
- Corregir los fallos de `adws` y el formato CRLF de Windows: son previos a esta rama (ver "Riesgos").
- Limpiar los planes duplicados de `plans/` ni los archivos sin seguimiento ajenos a la issue (`config/adw.yml`, `migrations/schema.sql`, `.claude/skills/ship/`).

## Diseño por capas

- **Persistencia**: columna `due_date` de tipo `date`, nullable. Migración ya creada y reversible.
- **Modelo**: `Task.DueDate *time.Time`. `NULL` = sin fecha. Sin hooks nuevos; `Validate` solo valida el título.
- **Acción y ruta**: `TasksCreate` y `TasksUpdate` parsean `due_date` con `parseDueDate`. Un error de formato se añade a `verrs` con clave `due_date`. Las rutas no cambian.
- **Renderizado**: `index` y `edit` muestran el input y la fecha con `Format`. El formato visible es `02/01/2006`; el del input es `2006-01-02`.
- **CSS**: `.task-due` y el ancho del date input en la fila de alta. Sin JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y quita la columna `due_date`. |
| `migrations/embed.go` | Embebe `*.fizz` y `*.sql` para las migraciones. |
| `models/task.go` | Campo `DueDate` y `Validate`. |
| `actions/tasks.go` | `parseDueDate`, `TasksCreate`, `TasksUpdate`. |
| `templates/tasks/index.plush.html` | Input de alta y fecha en la lista. |
| `templates/tasks/edit.plush.html` | Input de edición, ayuda y error. |
| `public/assets/app.css` | `.task-due` y `.new-task-row input[type="date"]`. |
| `models/task_test.go` | Tests de modelo para fecha opcional y persistencia de día. |
| `actions/tasks_test.go` | Tests de acción y plantilla para fecha. |
| `database.yml` | Configuración de `todo_test` (PostgreSQL en `127.0.0.1`). |
| `actions/app.go` | Middleware CSRF; se salta solo si `GO_ENV=test`. |

## Implementación paso a paso

1. Confirmar que el árbol de la rama está limpio respecto a los cambios de la issue: `git status --short`. Los archivos sin seguimiento y `config/adw.yml` no forman parte del PR.
2. Revisar el diff contra `main` (`git diff main...HEAD`) frente a los criterios 1–7. No modificar código salvo que una validación falle.
3. Ejecutar las validaciones de la sección siguiente con `GO_ENV=test` y registrar la salida.
4. Si todo pasa, no hay cambios de código. La rama ya contiene la implementación y sus tests; el paso siguiente es el PR, no un cambio nuevo.
5. Si una validación falla, parar y reportar la salida. No ampliar el alcance.

## Tests y validación

Requisitos del entorno:

- PostgreSQL escuchando en `127.0.0.1:5432` con usuario `postgres` / contraseña `postgres` y base `todo_test` (verificado: el puerto responde).
- **`GO_ENV=test`** es obligatorio para los tests de acción. Sin esa variable, el middleware CSRF (`actions/app.go`, `csrf.New`) responde 403 a los POST/PUT de los tests. Con `buffalo test` se define solo; con `go test` hay que ponerla.

Comandos y resultados observados en esta sesión (rama `feat/issue-3-add-task-due-date`):

- `GO_ENV=test go test -p 1 ./models/ ./actions/ -v`: **OK**. `Test_ModelSuite` 8/8 y `Test_ActionSuite` pasan, incluidos los 11 tests de fecha (2 de modelo y 9 de acción).
- Sin `GO_ENV=test`: `Test_ActionSuite` falla en 29 casos por 403 de CSRF. Lo mismo ocurre en `main` exportado a un directorio temporal (baseline), por tanto no es un fallo de esta rama.
- `go vet ./...`: sin salida, sin errores.
- `go build ./...`: sin errores.
- `gofmt -l` sobre el contenido versionado (`git show HEAD:<archivo> | gofmt -l`) de `actions/tasks.go`, `actions/tasks_test.go`, `models/task.go` y `models/task_test.go`: limpio.
- `GO_ENV=test go test -p 1 ./...`: `actions` y `models` OK; `adws` falla en 6 tests (`TestTrigger*`). El mismo fallo aparece en `main`; no se relaciona con la fecha.

Tests que demuestran el comportamiento (ya existentes; no hace falta añadir):

- Criterio 1: `Test_Task_DueDate_Optional`, `Test_TasksCreate_Valid`.
- Criterio 2: `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`.
- Criterio 3: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 4: `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- Criterio 5: `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`.
- Criterio 6: `Test_TasksIndex_ShowsDueDate`.
- Criterio 7: `Test_TasksToggle_KeepsDueDate`.

## Riesgos y recuperación

- **Falsos fallos de entorno**: sin `GO_ENV=test`, la suite de acciones parece rota. Documentado arriba; el baseline de `main` tiene el mismo comportamiento.
- **`gofmt -l .` da falsos positivos en Windows**: el árbol de trabajo tiene CRLF (`core.autocrlf=true`), y el índice guarda LF. Todos los `.go` del repositorio aparecen, también en `main`. Usar `git show HEAD:<archivo> | gofmt -l` para comprobar el contenido real.
- **Worktrees en `.claude/worktrees/`**: `gofmt -l .` también recorre sus copias. No son parte del PR.
- **Fallos de `adws`**: 6 tests de `TestTrigger*` fallan en `main`. Fuera de alcance; reportar sin corregir.
- **Pérdida de fecha en re-render**: ver "Huecos menores". Riesgo de UX, no de datos.
- **Zona horaria**: `time.Parse` produce medianoche UTC y la columna es `date`; mostrar con `Format` evita desplazamientos. Si se añade lógica de comparación con "hoy", revisar este punto.
- **Recuperación**: la migración tiene `down` (`drop_column`). Revertir la rama no requiere datos manuales.
- **Incertidumbre**: en una ejecución sin `GO_ENV=test` se vio la página de lista con 43 tareas en `todo_test`, aunque los tests crean sus propios datos. No se investigó la causa; con `GO_ENV=test` los tests de lista pasan.
