# Feature: fecha límite opcional (solo día) en las tareas (issue #3) — cierre de criterios

## Encargo y criterios de aceptación

Encargo de la issue #3 (título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"):

- Explícito: las tareas pueden tener una fecha límite.
- Explícito: la fecha es solo día, sin hora.
- Inferido: la fecha es opcional; una tarea sin fecha sigue siendo válida.
- Inferido: la fecha se crea, ve, edita y borra desde la interfaz existente (lista y edición de tareas).

Criterios de aceptación (los mismos del plan vigente `plans/issue-3-add-task-due-date-pendientes.md`):

1. Al crear una tarea, el formulario de la lista acepta una fecha opcional (`due_date`, tipo `date`).
2. Una tarea creada sin fecha se guarda con `due_date` nulo.
3. Una fecha válida `YYYY-MM-DD` se guarda como día sin hora y se muestra como `dd/mm/aaaa`.
4. Una fecha no válida no crea la tarea, muestra un error en `due_date`, y **el título y la fecha introducidos se conservan** en el formulario.
5. Un título vacío o demasiado largo se sigue rechazando, y la fecha introducida se conserva en el formulario.
6. Editar una tarea permite cambiar o borrar la fecha; un error de validación conserva el título y la fecha introducidos.
7. Completar o reabrir una tarea no modifica su fecha límite.

## Comportamiento actual y propuesta

Estado verificado leyendo el código de `feat/issue-3-add-task-due-date` (commit de referencia `9e6e3bc`). La funcionalidad ya está implementada; este plan no la reimplementa.

Cumplen hoy, con test asociado:

- Criterio 1 y 2: `TasksCreate` (`actions/tasks.go`), `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_Valid`.
- Criterio 3: `parseDueDate` y `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate` (`22/10/2026`).
- Criterio 5 (título inválido conserva la fecha válida): `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- Criterio 6 (cambiar y borrar): `Test_TasksUpdate_SetsAndClearsDueDate`. La edición con error de validación del título: `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- Criterio 7: `Test_TasksToggle_KeepsDueDate`.

**Hueco concreto (criterio 4 y parte del 6): una fecha inválida no se conserva.**

- `TasksCreate` y `TasksUpdate` hacen `task.DueDate = dueDate`. Si `parseDueDate` falla, `dueDate` es `nil`.
- Al volver a renderizar con error (422), `templates/tasks/index.plush.html` y `templates/tasks/edit.plush.html` pintan el `value` del input `date` a partir de `task.DueDate`. Con `nil` el campo queda vacío y el texto escrito por la persona se pierde.
- Ningún test comprueba que el valor inválido vuelva al formulario. `Test_TasksCreate_InvalidDueDate` solo mira el mensaje y el conteo de tareas. `Test_TasksUpdate_InvalidDueDate` solo mira el mensaje, `aria-describedby` y que la BD no cambia.

Propuesta: en la rama de error, devolver el texto tal como lo escribió la persona y pintarlo en el input.

- Acción: en `TasksCreate` y `TasksUpdate` pasar al render la cadena cruda `input.DueDate` (por ejemplo `c.Set("dueDateInput", input.DueDate)`).
- Para que Plush no falle con una variable indefinida, las rutas que renderizan los mismos templates también deben definirla: `TasksIndex` y `TasksEdit` con cadena vacía, y `TasksCreate`/`TasksUpdate` con `input.DueDate` solo en error.
- Plantillas: el `value` del input de alta y de edición usa `dueDateInput` si no está vacío; si está vacío, usa el formato de `task.DueDate` como hoy.
- Alternativa descartada: añadir un campo `DueDateInput` con `db:"-"` al modelo. Mezcla estado de formulario con la entidad; no sigue el patrón actual (`verrs` y `c.Set`).

Decisión sobre el nombre de la variable: `dueDateInput`, en camelCase como el resto de `c.Set` del repo (`tasks`, `task`, `errors`) no lo tiene; si el repo usa otra convención, el implementador debe seguirla. Verificar al implementar.

## Alcance

Incluido:

- Conservar el texto de la fecha inválida en el formulario de alta y de edición (criterio 4 y 6).
- Tests que demuestren ese comportamiento (ver sección de tests).
- Validar que la migración `add_due_date_to_tasks` sube y baja en la base de test, si el entorno lo permite.

Excluido:

- Reimplementar la fecha límite, la migración, el modelo o la lista ya existentes.
- Hora, zona horaria, recordatorios, filtros u orden por fecha.
- Cambios en `adws/`, en `config/adw.yml` (modificado en el working tree, no es de este cambio) o en `migrations/schema.sql`, `.txt`, `.claude/skills/ship/` (sin confirmar, no forman parte del PR).
- Los fallos de `adws/` que ya existen en `main`.
- Conversión de finales de línea CRLF: `gofmt -l .` marca casi todos los archivos por el entorno Windows, no por este cambio.

## Diseño por capas

- **Persistencia**: sin cambios. Columna `due_date` `date` nullable ya migrada.
- **Modelo**: sin cambios. `DueDate *time.Time`; la validación de formato vive en la acción (`parseDueDate`), como ya está.
- **Acción y ruta**: `TasksIndex`, `TasksCreate`, `TasksEdit`, `TasksUpdate` pasan `dueDateInput` al render. `TasksToggle` y `TasksDestroy` no cambian.
- **Renderizado**: `templates/tasks/index.plush.html` (input de alta) y `templates/tasks/edit.plush.html` (input de edición) usan `dueDateInput` cuando no está vacío.
- **CSS**: sin cambios.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `actions/tasks.go` | `parseDueDate`, render de alta y edición con la cadena cruda en error |
| `actions/tasks_test.go` | Tests de conservación del valor inválido en alta y edición |
| `templates/tasks/index.plush.html` | `value` del input `due_date` del formulario de alta |
| `templates/tasks/edit.plush.html` | `value` del input `due_date` del formulario de edición |
| `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` | Solo lectura para la validación up/down |
| `models/task.go`, `models/task_test.go` | Sin cambios |

## Implementación paso a paso

1. Partir de la rama `feat/issue-3-add-task-due-date` tal como está. No tocar los archivos sin confirmar que no son de este cambio.
2. Añadir tests que fallen con el código actual:
   - `Test_TasksCreate_InvalidDueDate`: además del mensaje, comprobar `value="no-es-fecha"` en la respuesta.
   - `Test_TasksUpdate_InvalidDueDate`: comprobar `value="no-es-fecha"` en la respuesta.
3. Ejecutar los tests nuevos y comprobar que fallan por la razón esperada (el valor no aparece). Si pasan ya, el hueco no existe y se informa.
4. En `actions/tasks.go`: pasar `dueDateInput` en los cuatro handlers de render (`TasksIndex`, `TasksCreate` en error, `TasksEdit`, `TasksUpdate` en error). Mantener el orden y el estilo de los `c.Set` existentes.
5. En `templates/tasks/index.plush.html` y `templates/tasks/edit.plush.html`: el `value` del input `due_date` usa `dueDateInput` si no está vacío; si no, el formato de `task.DueDate` como hoy.
6. Ejecutar los tests de `actions/` y `models/` hasta que pasen.
7. Si el entorno lo permite, comprobar la migración en `todo_test` con `soda migrate -e test` y revertir la última con `soda migrate down -e test`; volver a aplicarla. Si no es posible, dejarlo indicado en el reporte.
8. No tocar archivos fuera de la lista de archivos relevantes.

## Tests y validación

Condiciones necesarias:

- PostgreSQL local escuchando en `127.0.0.1:5432` (comprobado: el puerto está abierto), con la base `todo_test` y usuario `postgres`/`postgres`, y las migraciones aplicadas (`soda migrate -e test`).
- `GO_ENV=test` en el entorno. Sin esta variable, `actions/` falla con 403 por CSRF; CI la define en `.github/workflows/ci.yaml`.

Comprobaciones:

- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/` debe pasar. Antes de implementar, los dos tests nuevos deben fallar.
- `go vet ./...` y `go build ./...` deben pasar.
- `gofmt -l` sobre los archivos Go modificados (`actions/tasks.go`, `actions/tasks_test.go`) debe quedar vacío, o explicar cualquier salida que sea solo CRLF. No usar `gofmt -l .` como criterio.
- Los fallos de `adws/` existen también en `main` y no son regresión de esta feature.
- Migración up/down: opcional según el punto 7 del paso a paso.

Tests que demuestran el comportamiento:

- Ya presentes: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksToggle_KeepsDueDate`, `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
- Nuevas: aserciones de `value` con la fecha inválida en los dos tests de error del paso 2. Esto demuestra el criterio 4 y la parte de edición del 6.

## Riesgos y recuperación

- Variable indefinida en Plush: si alguna ruta de render no define `dueDateInput`, el template falla en tiempo de ejecución. Por eso se define en las cuatro rutas; los tests de `Test_TasksIndex_*` y `Test_TasksEdit_*` lo cubren.
- Escapado HTML: la cadena cruda se pinta con `<%= %>` (escapado automático de Plush). Un valor como `"><script>` no debe romper el atributo; comprobarlo solo si el implementador lo considera relevante.
- Zona horaria: `time.Parse` sin zona produce medianoche UTC. Al formatear sin conversión el día no cambia. Lo cubre `Test_Task_DueDate_DayOnly_Persists`.
- Rollback: revertir el cambio de `actions/tasks.go` y las plantillas vuelve al comportamiento actual (fecha inválida no conservada, pero funcional). No hay cambio de esquema en este plan.
- Incertidumbre 1: el nombre `dueDateInput` es propuesta; confirmar la convención del repo al implementar.
- Incertidumbre 2: no se ha ejecutado la suite en esta sesión; el estado "pasa" del plan vigente es una afirmación previa que el implementador debe reconfirmar.
- Incertidumbre 3: no se ha ejecutado la migración `down`; el paso 7 la deja condicionada al entorno.
