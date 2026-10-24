# Feature: Fecha límite de las tareas, solo día (issue #3)

## Encargo y criterios de aceptación

**Encargo (issue #3, "añadir fechaimie alas task", estado OPEN):** las tareas deben poder tener una fecha límite. Comentario de la issue: "vale un dia sin hora", es decir, la fecha es solo día, sin hora.

**Requisitos explícitos (de la issue):**

- R1. Una tarea puede tener una fecha límite.
- R2. La fecha límite es solo día, sin hora.

**Inferencias (no escritas en la issue; se asumen para cerrarla):**

- I1. La fecha es opcional: una tarea sin fecha sigue siendo válida (la creación actual sin fecha no debe romperse).
- I2. La fecha se introduce y se edita desde los formularios existentes (alta en la lista y edición de título), usando `<input type="date">`.
- I3. La fecha se muestra en la lista de tareas en formato `dd/mm/aaaa`.
- I4. Una fecha inválida se rechaza con un mensaje en español y no crea ni actualiza la tarea.
- I5. Completar o reabrir una tarea conserva su fecha límite.
- I6. Editar solo el título conserva la fecha límite; quitar la fecha se hace enviando el campo vacío.

**Criterios de aceptación (verificables):**

- CA1. Crear una tarea sin fecha deja `due_date` nulo en la base de datos (`Test_TasksCreate_Valid`).
- CA2. Crear una tarea con `due_date=2026-10-22` persiste exactamente ese día, con hora 00:00 (`Test_TasksCreate_WithDueDate`, `Test_Task_DueDate_DayOnly_Persists`).
- CA3. Una fecha inválida devuelve 422 con "La fecha no es válida." y no crea la tarea (`Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`).
- CA4. Un título vacío con fecha válida devuelve 422 y conserva la fecha en el formulario (`Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`).
- CA5. La lista muestra "Fecha límite: 22/10/2026" para una tarea con fecha (`Test_TasksIndex_ShowsDueDate`).
- CA6. La edición precarga la fecha existente en `value="2026-10-22"` (`Test_TasksEdit_PrefillsDueDate`).
- CA7. Actualizar y luego borrar la fecha la deja nula; el título y el estado `Completed` no cambian (`Test_TasksUpdate_SetsAndClearsDueDate`).
- CA8. Editar solo el título conserva la fecha (`Test_TasksUpdate_KeepsDueDateWhenOnlyTitleChanges`).
- CA9. Completar/reabrir conserva la fecha (`Test_TasksToggle_KeepsDueDate`).

## Comportamiento actual y propuesta

**Estado actual de la rama `feat/issue-3-add-task-due-date` (verificado leyendo el diff contra `main`, no ejecutando tests):**

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `.down.fizz` que hace `drop_column`. Tipo `date`, por lo que la hora no existe en BD.
- `models/task.go`: campo `DueDate *time.Time` con `json:"due_date" db:"due_date" form:"due_date"`. Puntero para representar "sin fecha".
- `actions/tasks.go`:
  - `parseDueDate(raw string)` convierte `YYYY-MM-DD` con `time.Parse("2006-01-02", ...)`; cadena vacía devuelve `nil`.
  - `TasksCreate` y `TasksUpdate` leen `title` y `due_date`, construyen la tarea, llaman a `task.Validate(tx)`, añaden el error `due_date` si el parseo falla y, si no hay errores, `tx.Create` / `tx.Update`.
  - `TasksEdit` y `TasksIndex` fijan `dueDateInput` para repintar el valor enviado.
  - `TasksUpdate` solo cambia título y fecha; `Completed`, `CreatedAt` y `ID` se conservan porque se cargan de la BD antes de actualizar.
- `templates/tasks/index.plush.html`: input `type="date"` `name="due_date"` en el formulario de alta, span `.task-due` con `Fecha límite: dd/mm/aaaa` y mensaje de error `due_date`.
- `templates/tasks/edit.plush.html`: campo `#task-due-date` con ayuda `#due-date-help`, error `#due-date-error` y `aria-invalid` cuando hay errores.
- `public/assets/app.css`: clases `.task-due` y `.new-task-row input[type="date"]`.
- `actions/actions_test.go`: helper `authed()` que añade el token CSRF a las peticiones que cambian estado y helper `csrfToken()` que lo lee del HTML.
- Tests añadidos en `actions/tasks_test.go`, `models/task_test.go` y `actions/actions_test.go` (lista en CA1–CA9).

**Propuesta:** no implementar nada nuevo en código de aplicación. Cerrar la feature verificando que la implementación existente cumple R1, R2 e I1–I6 y que sus tests pasan. El cambio de esta ejecución se limita a:

1. Confirmar, ejecutando las validaciones, que el código de la rama cumple los criterios CA1–CA9.
2. Si algún test falla por causas de la feature, corregir solo ese defecto dentro del alcance (ver "Riesgos").
3. Preparar el PR con los archivos de la feature y este plan.

## Alcance

**Incluido:**

- Verificación de migración, modelo, acciones, plantillas, CSS y tests ya presentes en la rama.
- Ejecución de las validaciones del repositorio y corrección de fallos directos de esta feature.
- Un único plan nuevo en `plans/` (este archivo) como artefacto del cambio.

**Excluido:**

- Hora de la fecha, recordatorios, notificaciones, filtros por fecha o orden por fecha límite (no pedidos).
- Cambios en `config/adw.yml` (modificación local no relacionada, ya presente en el árbol de trabajo).
- Archivos sin seguimiento que no son de esta feature: `.claude/skills/ship/`, `.txt`, `.x`, `migrations/schema.sql`.
- Los más de 160 planes previos de issue-3 en `plans/` (ver "Planes previos obsoletos").
- Cambios de dependencias o de configuración de base de datos.

## Diseño por capas

- **Persistencia y migraciones:** ya existe. Columna `due_date date NULL` en `tasks`, con migración reversible. Se valida que `fizz` la aplica y revierte en `todo_test`.
- **Modelo:** ya existe. `DueDate *time.Time` con etiquetas `db`, `json`, `form`. Sin cambios necesarios.
- **Acción y ruta:** ya existen. Rutas `POST /tasks`, `PUT /tasks/{task_id}` y `PUT /tasks/{task_id}/toggle` en `actions/app.go`. No se añaden rutas.
- **Renderizado (Plush):** ya existe en `index` y `edit`.
- **CSS:** ya existe; solo clases de presentación.
- **JavaScript:** no aplica. El formulario usa envío HTML normal.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date date NULL`. |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna. |
| `models/task.go` | Campo `DueDate *time.Time`. |
| `models/task_test.go` | Pruebas de modelo: opcional y persistencia solo día. |
| `actions/tasks.go` | `parseDueDate`, alta, edición y actualización con fecha. |
| `actions/tasks_test.go` | Pruebas de comportamiento HTTP de la fecha. |
| `actions/actions_test.go` | Helpers `authed` y `csrfToken` para peticiones con CSRF. |
| `actions/app.go` | Rutas existentes (sin cambios). |
| `templates/tasks/index.plush.html` | Input de fecha en alta, visualización en lista, error. |
| `templates/tasks/edit.plush.html` | Campo de fecha en edición, ayuda y error. |
| `public/assets/app.css` | Estilo de `.task-due` y del input de fecha. |

## Implementación paso a paso

1. Comprobar el estado de la rama: `git status --short --branch` y `git log --oneline main..HEAD`. Confirmar que el árbol contiene los commits de la feature y que `config/adw.yml` sigue siendo el único cambio tracked fuera de la feature.
2. Confirmar `GO_ENV=test` y la base `todo_test` disponible (ver "Tests y validación").
3. Aplicar las migraciones en `todo_test` (`buffalo test` o el comando del repositorio) y comprobar que `due_date` existe.
4. Ejecutar las validaciones (ver "Tests y validación") y registrar la salida.
5. Si todo pasa, no tocar código. Si falla un test de la feature, corregir la causa mínima dentro de los archivos listados arriba y repetir las validaciones.
6. Preparar el commit/PR solo con los archivos de la feature y este plan (ver "Riesgos").

## Tests y validación

**Condiciones previas:**

- PostgreSQL accesible en `127.0.0.1:5432` (o `DB_HOST`) con usuario `postgres`/`postgres`.
- Base `todo_test` creada y con las migraciones aplicadas.
- `GO_ENV=test` en el entorno. Sin él, `actions/` y `models/` leen la configuración de `development` y escriben en `todo_development`. Ver la nota de memoria `go-tests-need-go-env-test`.

**Comandos:**

- `gofmt -l .` — debe devolver vacío.
- `go vet ./...` — debe terminar sin errores.
- `go build ./...` — debe terminar sin errores.
- `GO_ENV=test go test -p 1 ./...` — todos los tests en verde; `-p 1` evita carreras sobre la misma base.

**Tests que demuestran el comportamiento:** los listados en CA1–CA9. Todos están ya en la rama. Ninguno se ha ejecutado todavía en esta planificación.

**Cobertura de la validación:** no hay tests nuevos que añadir salvo que la ejecución revele un hueco. Si la ejecución falla por un caso no cubierto (p. ej. una fecha como `2026-02-30`), se añade un test de acción junto a `Test_TasksCreate_InvalidDueDate`, siguiendo el mismo patrón.

## Riesgos y recuperación

- **Tests no ejecutados:** la afirmación "la feature funciona" depende de la ejecución. Si PostgreSQL o `todo_test` no están disponibles, el plan se detiene y se informa; no se abre PR.
- **GO_ENV olvidado:** los tests escribirían en `todo_development`. Verificar `GO_ENV=test` antes de ejecutar.
- **Fecha en zona horaria:** `time.Parse` produce medianoche UTC y la columna es `date`. Los tests comparan con `Format("2006-01-02")`, que es estable. Riesgo bajo.
- **Archivos ajenos en el PR:** `config/adw.yml`, `.claude/skills/ship/`, `.txt`, `.x` y `migrations/schema.sql` no son de esta feature. Excluirlos del stage (lista explícita, nunca `git add .`).
- **Planes previos obsoletos:** los ~164 archivos `plans/issue-3-*` y `plans/*fecha-limite*` / `plans/verificar-fecha-limite-tareas.md` son borradores de ejecuciones anteriores, en su mayoría "verificado" o "cierre". Quedan obsoletos frente a este plan. No se borran ni se commitean en este cambio; se listan como pendiente de limpieza.
- **Recuperación:** la migración es reversible con su `.down.fizz`. Si el PR se revierte, `git revert` de los commits de la feature devuelve `main` sin la columna tras ejecutar la migración de bajada.

## Incertidumbres pendientes

- Si las validaciones pasan con el estado actual de la rama. Hasta ejecutarlas, el cierre es una hipótesis basada en la lectura del diff.
- Si la issue #3 debe cerrarse al mergear el PR (`Closes #3` en el cuerpo) o solo referenciarse. Por defecto se usa `Closes #3`.
- Si los commits existentes de la rama bastan para el PR o si conviene reagrupar en un único commit de cierre. Por defecto se mantienen los existentes.

## Siguiente paso

`/implement plans/issue-3-fecha-limite-solo-dia-plan-cierre-feature.md`
