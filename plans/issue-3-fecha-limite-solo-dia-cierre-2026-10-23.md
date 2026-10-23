# Feature: fecha límite opcional (solo día) en las tareas — cierre de la issue #3

## Encargo y criterios de aceptación

Encargo (issue #3, `añadir fechaimie alas task`, cuerpo: `vale un dia sin hora`):

- Explícito: cada tarea puede tener una fecha límite opcional.
- Explícito: la fecha es solo día, sin hora ("vale un dia sin hora").
- Inferido: la fecha se puede definir al crear y editar, se muestra en la lista y se puede borrar.
- Inferido: completar o reabrir una tarea no altera su fecha límite.

Criterios de aceptación:

1. Crear una tarea sin fecha funciona como antes (`due_date` NULL).
2. Crear una tarea con `YYYY-MM-DD` guarda ese día sin hora.
3. Una fecha con formato inválido no crea la tarea y muestra el error en `due_date`, conservando el título y la fecha escrita.
4. Editar permite fijar, cambiar y borrar la fecha límite.
5. La lista muestra `Fecha límite: DD/MM/AAAA` solo cuando hay fecha.
6. Un título inválido al crear o editar conserva la fecha introducida en el formulario.

## Comportamiento actual y propuesta

Estado de la rama `feat/issue-3-add-task-due-date` frente a `main` (verificado en el código, no en planes previos):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: columna `due_date` tipo `date`, nullable.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` (vacío → nil, `time.Parse("2006-01-02")`), validación de `due_date` en `TasksCreate` y `TasksUpdate`, persistencia con `tx.Create` / `tx.Update` tras `task.Validate(tx)`, `dueDateInput` para repintar el valor tras error.
- Plantillas: `templates/tasks/index.plush.html` (input date en el alta y `Fecha límite: DD/MM/AAAA` en la lista), `templates/tasks/edit.plush.html` (campo de edición con ayuda y error).
- `public/assets/app.css`: clases `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Tests: `actions/tasks_test.go` (≈129 líneas nuevas: alta con/sin fecha, fecha inválida, título inválido conservando fecha, edición fijar/borrar, edición con error, edición del formulario, toggle conserva fecha, lista muestra fecha), `actions/actions_test.go` (helpers `authed` y `csrfToken` para CSRF), `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`).
- PR abierto: #5 (`OPEN`) desde esta misma rama.

Propuesta: no hay cambio de comportamiento nuevo. El trabajo de esta sesión es cerrar la rama tal cual está: verificar, confirmar que los criterios se cumplen con los tests existentes y actualizar el PR #5 con los commits necesarios. Si la validación detecta un fallo real de la feature, se corrige solo ese fallo dentro de este mismo alcance.

## Alcance

Incluido:

- Verificación de los criterios 1–6 contra los tests y el código existentes.
- Ejecución de la validación (ver más abajo) y registro de resultados.
- Subida de la rama a `origin` para actualizar el PR #5, si la validación pasa.

Excluido:

- Nuevos campos, filtros, ordenación por fecha, recordatorios, zona horaria o hora.
- Refactors de `actions/tasks.go` o de los templates que no sean necesarios para un fallo confirmado.
- Cambios en `adws/` (ver "Tests y validación").
- Los ficheros sin confirmar que no pertenecen a la feature: `config/adw.yml` (modificado), `migrations/schema.sql`, `.txt`, `.x`, `.claude/skills/ship/`, y los `plans/` de iteraciones anteriores (ver "Riesgos y recuperación").

## Diseño por capas

- **Persistencia y migración**: ya implementada. Columna `date` nullable; no se añade hora, por eso el tipo es `date` y no `timestamp`.
- **Modelo**: ya implementado. `*time.Time` representa "sin fecha" con `nil`.
- **Acción y ruta**: ya implementadas. Las rutas `/tasks` y `/tasks/{id}` no cambian. La validación de la fecha se hace en la acción, no en el modelo, porque el error de formato solo existe en la entrada del formulario.
- **Renderizado**: ya implementado. Repinta `dueDateInput` tras error para no perder lo escrito.
- **CSS**: ya implementado, solo para el input de fecha y la etiqueta de la lista.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade `due_date` date nullable |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna |
| `models/task.go` | Campo `DueDate *time.Time` |
| `actions/tasks.go` | Parseo, validación y persistencia de la fecha en alta y edición |
| `templates/tasks/index.plush.html` | Input de fecha en el alta y etiqueta en la lista |
| `templates/tasks/edit.plush.html` | Campo de fecha en edición con error y ayuda |
| `public/assets/app.css` | Estilos `.task-due` y del input de fecha |
| `actions/tasks_test.go` | Tests de comportamiento de alta, edición, lista y toggle |
| `actions/actions_test.go` | Helpers `authed` y `csrfToken` para peticiones con CSRF |
| `models/task_test.go` | Tests del campo opcional y de persistencia del día |

## Implementación paso a paso

No hay implementación pendiente. Pasos para cerrar:

1. Confirmar que la rama contiene todos los commits de la feature y que no hay cambios de aplicación sin confirmar (`git status --short -- actions models migrations templates public`). Debe estar vacío.
2. Ejecutar la validación de la sección siguiente y guardar la salida.
3. Si todo pasa, hacer stage solo de los ficheros de la feature y del plan (ver lista abajo), commit solo si hay cambios, y `git push origin feat/issue-3-add-task-due-date` (sin `--force`). El PR #5 se actualiza solo.
4. Actualizar el título o la descripción del PR #5 solo si difieren del resumen de la feature. No abrir un PR nuevo.

Ficheros que entrarían en un commit de cierre, si fuera necesario: el plan de esta carpeta y cualquier fichero de la lista de "Archivos relevantes" que se modifique. Nada más.

## Tests y validación

Validaciones pertinentes:

- `gofmt -l actions models migrations`: en este repositorio los ficheros tienen finales de línea CRLF en el disco, y `gofmt` los marca todos aunque su contenido esté bien formateado. Antes de decidir, comprobar con `gofmt -d` que la única diferencia es el CRLF. No ejecutar `gofmt -l .` sin excluir `.claude/worktrees/`, porque incluye copias de otras ramas.
- `go vet ./...`: pasa.
- `go build ./...`: pasa.
- `GO_ENV=test go test -p 1 ./actions/... ./models/...`: pasan (la salida previa estaba en caché; repetir con `-count=1` antes de subir).
- `GO_ENV=test go test -p 1 ./adws/...`: falla en 6 tests porque consultan el repositorio `example/todo` vía GraphQL. Como `adws/` no tiene diferencias con `main` en esta rama, no es un fallo de la feature. Se reporta tal cual y no se arregla en este alcance.

Test que demuestra cada criterio:

- Criterio 1: `Test_TasksCreate_WithDueDate` (rama sin fecha) y `Test_Task_DueDate_Optional`.
- Criterio 2: `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`.
- Criterio 3: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 4: `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`.
- Criterio 5: `Test_TasksIndex_ShowsDueDate`.
- Criterio 6: `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- Inferido, completar conserva fecha: `Test_TasksToggle_KeepsDueDate`.

## Riesgos y recuperación

- **Cambios sin confirmar fuera de la feature**: `config/adw.yml`, `migrations/schema.sql`, `.txt`, `.x`, `.claude/skills/ship/` y decenas de planes en `plans/`. No forman parte de esta feature; no deben entrar en el commit ni en el PR. El flujo `ship` pide avisar de ellos, y así se hace en el reporte.
- **Falsos positivos de gofmt por CRLF**: si el CRLF se normaliza en un commit, el diff de todos los ficheros de Go cambia. No convertir líneas en este cierre.
- **Migración ya aplicada en bases compartidas**: `20261022120000` añade una columna nullable; revertir con el `.down.fizz` la elimina y pierde los datos de fecha. No aplicar ni revertir en bases de datos que no sean de test.
- **Dependencia de `adws`**: sus tests no son parte del PR; si el CI los ejecuta, el PR fallará por motivos ajenos a la feature. Señalarlo al revisor.
- **Plan duplicado**: ya existen más de 100 planes sobre esta issue en `plans/`. Este documento sustituye a los de cierre previos; no crear otro en el siguiente intento.
