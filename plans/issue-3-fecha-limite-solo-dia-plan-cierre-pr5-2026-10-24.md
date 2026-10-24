# Feature: fecha límite opcional (solo día) en las tareas, cierre del PR #5 (issue #3)

## Encargo y criterios de aceptación

Encargo de la issue #3 ("añadir fechaimie alas task"), con el comentario "vale un dia sin hora": cada tarea puede tener una fecha límite opcional que indica solo el día, sin hora.

Criterios de aceptación:

1. Crear una tarea permite indicar una fecha límite opcional (`YYYY-MM-DD`). Sin fecha, la tarea se crea igual que antes.
2. Editar una tarea permite fijar, cambiar o borrar la fecha límite. Borrarla deja la columna a `NULL`.
3. La fecha se guarda sin hora. La lista y el formulario muestran solo el día.
4. Una fecha no válida se rechaza con un mensaje, no se guarda y el valor introducido se conserva en el formulario.
5. Si falla el título en alta o edición, la fecha introducida se conserva en el formulario.
6. Completar o reabrir una tarea conserva su fecha límite.
7. Las tareas existentes, sin fecha, siguen funcionando.

### Requisitos explícitos e inferencias

- Explícito (issue): fecha límite en las tareas.
- Explícito (comentario): solo día, sin hora.
- Inferido: campo opcional y nulo en base de datos, porque las tareas existentes no tienen fecha.
- Inferido: formato `YYYY-MM-DD` en el formulario, que es el que produce el `<input type="date">` nativo.
- Inferido: visualización `DD/MM/YYYY` en la lista, consistente con el resto de la interfaz en español.

## Comportamiento actual y propuesta

Estado verificado leyendo el código de `feat/issue-3-add-task-due-date` (HEAD `c1e0e22`, con la implementación en `1fc855a` y `d222c26`). El PR #5 (base `main`, abierto) ya contiene esa rama.

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. El `down` la elimina.
- `models/task.go`: `Task.DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` solo valida el título.
- `actions/tasks.go`:
  - `parseDueDate` convierte `""` en `nil` y `YYYY-MM-DD` en `*time.Time`; otro valor devuelve error.
  - `TasksCreate` y `TasksUpdate` leen `title` y `due_date` en el mismo struct, añaden `due_date` a los errores si el parseo falla (`dueDateInvalidMessage`), y si hay errores re-renderizan con `dueDateInput` = valor crudo. Si no, persisten con `tx.Create` o `tx.Update`.
  - `TasksIndex` y `TasksEdit` fijan `dueDateInput` a `""`.
  - `TasksToggle` solo cambia `Completed` sobre la tarea leída de la base, así que la fecha se conserva.
- `templates/tasks/index.plush.html`: input `type="date"` `name="due_date"` en el alta; bloque `task-due` con `DD/MM/YYYY` en las tareas con fecha.
- `templates/tasks/edit.plush.html`: input `type="date"` con etiqueta, ayuda, `aria-invalid`/`aria-describedby` y mensaje de error.
- `public/assets/app.css`: clase `.task-due` y ancho del input de fecha en `.new-task-row`.

Contraste con los criterios (lectura del código y de los tests):

| Criterio | Dónde se cumple | Test que lo demuestra |
|---|---|---|
| 1 | `TasksCreate`, `parseDueDate("")` → nil | `Test_TasksCreate_Valid`, `Test_TasksCreate_WithDueDate` |
| 2 | `TasksUpdate` escribe `DueDate` (nil → NULL) | `Test_TasksUpdate_SetsAndClearsDueDate` |
| 3 | Columna `date`, `Format("02/01/2006")` en lista | `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksIndex_ShowsDueDate` |
| 4 | `verrs.Add("due_date", …)` y `dueDateInput` | `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate` |
| 5 | `dueDateInput` en el re-render de error | `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate` |
| 6 | `TasksToggle` no toca `DueDate` | `Test_TasksToggle_KeepsDueDate` |
| 7 | Columna nullable | `Test_Task_DueDate_Optional`, `Test_TasksUpdate_PersistsTitleOnly` |

Conclusión: no hay huecos funcionales. La implementación cubre los criterios 1–7 y cada uno tiene al menos un test que fallaría si se rompiera.

Propuesta: no hay cambios de comportamiento ni de código. Falta confirmar la validación en el estado actual y actualizar el PR #5 con la descripción y el resultado.

## Alcance

Incluido:

- Confirmar que el árbol de trabajo no tiene cambios pendientes en las rutas del cambio.
- Ejecutar la validación de la sección siguiente.
- Actualizar la descripción del PR #5 (`gh pr edit 5 --body-file …`) con resumen, plan, tests y resultado.

Excluido:

- Hora de vencimiento, recordatorios, ordenación o filtro por fecha, y zonas horarias.
- Cualquier cambio en `actions/`, `models/`, `migrations/`, `templates/` o `public/`. Si la validación falla, no se corrige aquí: se informa y se detiene (ver "Riesgos y recuperación").
- Cambios en `adws/`, `config/adw.yml` (modificado sin relación con este cambio) y los archivos sin seguimiento (`.txt`, `.x`, `migrations/schema.sql`, `.claude/skills/ship/`, `plans/` sueltos).
- Crear un PR nuevo: ya existe el #5 para esta rama.
- Reescribir o borrar los planes previos de `plans/issue-3-*`.

## Diseño por capas

Solo se verifica; no hay capas que cambiar. Las capas revisadas son persistencia (migración), modelo, acción y ruta, renderizado y CSS. JavaScript no aplica: el formulario usa el input nativo de fecha.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade `due_date` nullable |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna |
| `models/task.go` | Campo `DueDate` |
| `models/task_test.go` | Fecha opcional y persistencia solo día |
| `actions/tasks.go` | `parseDueDate`, alta, edición, re-renders con error |
| `actions/tasks_test.go` | Alta, edición, errores, lista, prefill, completar |
| `templates/tasks/index.plush.html` | Input de alta y fecha en la lista |
| `templates/tasks/edit.plush.html` | Input de edición con errores accesibles |
| `public/assets/app.css` | `.task-due` y ancho del input de fecha |

## Implementación paso a paso

No hay código que escribir. Pasos:

1. Comprobar que no hay cambios sin confirmar en las rutas del cambio: `git status --short -- actions models migrations templates public`. Debe salir vacío.
2. Ejecutar la validación de la sección siguiente, en este orden. Si algo falla, parar e informar con la salida; no abrir ni actualizar el PR.
3. Si la validación pasa, comprobar `git status -sb`. Si la rama va por delante de `origin`, hacer `git push origin feat/issue-3-add-task-due-date` (sin `--force`). Si ya está al día, no hace falta push.
4. Actualizar el PR #5 con `gh pr edit 5 --repo clerma-arch/desarrollo-agentico-todo-03 --body-file <archivo>`. No ejecutar `gh pr create`.
5. No hacer merge.

## Tests y validación

Tests que demuestran el comportamiento (ya existen; no hace falta añadir ninguno):

- `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate`.
- `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.

Comandos (desde la raíz del repositorio):

```
gofmt -l .
go vet ./...
GO_ENV=test go test -count=1 -p 1 ./...
go build ./...
```

Notas de entorno:

- `GO_ENV=test` es obligatorio para `actions/`: usan la base `todo_test`. Sin él, la app escribe en la base de desarrollo.
- Los tests de `actions/` necesitan PostgreSQL con la base de test y las migraciones aplicadas. Si no está disponible, la validación no puede darse por pasada: parar y reportarlo.
- `gofmt -l .` puede marcar casi todos los archivos por los finales de línea CRLF del árbol de trabajo (`core.autocrlf=true`; el índice tiene LF). No es un fallo de formato. Para los archivos del cambio, comprobar con: `tr -d '\r' < actions/tasks.go | gofmt -l` (y lo mismo con cada archivo Go modificado). Salida esperada: vacía.
- Los tests de `adws/` dependen de `gh` y de un repositorio real. Un plan anterior registró que fallan en este entorno por `GraphQL: Could not resolve to a Repository`. Este cambio no toca `adws/`; si fallan, se documenta en el PR como fallo preexistente del entorno, pero se verifica antes de darlo por hecho.

Criterio de paso: `gofmt` sin salida útil para los archivos del cambio, `go vet` y `go build` sin errores, y `actions` y `models` en `ok`.

## Riesgos y recuperación

- **PR duplicado.** Ya existe el #5. Mitigación: actualizar ese PR, no crear otro.
- **Validación en rojo.** Si `actions` o `models` fallan, el PR no se toca y se informa con la salida. La corrección sería un cambio de código, fuera del alcance de este plan.
- **Migración en un entorno sin la columna.** Recuperación: el `down` elimina `due_date`; los valores de fecha se pierden y eso es irreversible.
- **Zonas horarias.** La fecha se parsea como medianoche UTC y se formatea sin conversión, así que un cambio de zona del servidor no la desplaza. Si algún día se añade hora, habrá que revisarlo.
- **Archivos sin seguimiento y `config/adw.yml`.** No forman parte del cambio y no deben entrar en el PR.
- **`adws` en rojo.** Si la CI ejecuta `./...`, el fallo puede aparecer. Mitigación: documentarlo en la descripción del PR.
