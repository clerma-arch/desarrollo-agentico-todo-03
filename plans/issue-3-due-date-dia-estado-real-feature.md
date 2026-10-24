# Feature: fecha límite de tareas, solo día (issue #3)

## Encargo y criterios de aceptación

Encargo de la issue #3 de `clerma-arch/desarrollo-agentico-todo-03` (título: "añadir fechaimie alas task"; cuerpo: "vale un dia sin hora"):

Requisitos explícitos:

- R1. Las tareas pueden tener una fecha límite.
- R2. La fecha límite es solo un día, sin hora.

Inferencias (no están en la issue, se toman por coherencia con lo ya implementado en la rama):

- I1. La fecha límite es opcional; una tarea sin fecha sigue siendo válida.
- I2. La fecha se puede definir al crear y cambiar o borrar al editar.
- I3. Una fecha con formato inválido no se guarda y se muestra un error en el formulario.
- I4. La fecha se muestra en la lista en formato `dd/mm/yyyy` y en el formulario de edición en `yyyy-mm-dd` (formato nativo de `input type="date"`).
- I5. Completar o reabrir una tarea conserva su fecha límite.

Criterios de aceptación observables:

- CA1. Crear una tarea con `due_date=2026-10-22` la guarda con ese día; la lista la muestra como `Fecha límite: 22/10/2026`.
- CA2. Crear una tarea sin `due_date` la guarda con `due_date` nulo y no muestra la etiqueta.
- CA3. Un `due_date` no válido (p. ej. `no-es-fecha`) responde 422, no crea la tarea y muestra el error ligado al campo (`aria-describedby="due-date-error"`).
- CA4. Un título vacío con fecha válida responde 422 y conserva la fecha en el formulario.
- CA5. Editar una tarea permite fijar, cambiar y borrar la fecha; con fecha inválida no actualiza la tarea.
- CA6. El formulario de edición precarga la fecha existente.
- CA7. Completar/reabrir (`TasksToggle`) no altera la fecha.
- CA8. La migración añade la columna `due_date` de tipo `date` nullable, con rollback limpio.

## Comportamiento actual y propuesta

Estado actual de la rama `feat/issue-3-add-task-due-date` (27 commits sobre `main`, ya publicados en `origin` según `git status -sb`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: `add_column("tasks", "due_date", "date", {"null": true})` y su `drop_column`. Cumple CA8.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. Sin cambios de validación en el modelo (la fecha inválida se valida en la acción).
- `actions/tasks.go`:
  - `parseDueDate` convierte `YYYY-MM-DD` a `*time.Time` (UTC a medianoche); vacío → `nil`; error si el formato es inválido.
  - `TasksIndex`, `TasksCreate`, `TasksEdit`, `TasksUpdate` usan la fecha; los errores de fecha se añaden a `verrs` bajo la clave `due_date` con el mensaje `dueDateInvalidMessage`.
  - `TasksToggle` no toca el campo (la fecha viaja con el registro cargado).
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de alta, etiqueta `Fecha límite: dd/mm/yyyy` bajo el título cuando existe, y error `due-date-error`.
- `templates/tasks/edit.plush.html`: input `type="date"` con `id="task-due-date"`, ayuda `due-date-help` y error `due-date-error`.
- `public/assets/app.css`: clase `.task-due` y ajuste de ancho para `input[type="date"]` dentro de `.new-task-row`.
- Tests en `actions/tasks_test.go`, `actions/actions_test.go` y `models/task_test.go` que cubren CA1–CA7 (crear, inválido, título inválido con fecha, editar fijar/borrar/inválido, precarga, toggle, render de la lista).

Propuesta: no cambiar el diseño. Verificar que el código de la rama cumple R1, R2 y CA1–CA8, cerrar lo que falte y dejar la rama lista para PR.

Brecha conocida: en la sesión de planificación no se ha ejecutado la suite de tests de aplicación; la verificación de ejecución queda pendiente (ver "Tests y validación").

## Alcance

Incluido:

- Verificación de la implementación existente frente a R1, R2, I1–I5 y CA1–CA8.
- Cualquier ajuste mínimo que resulte de esa verificación (solo si un criterio no se cumple).
- Ejecución de las validaciones y commit/push de los cambios del propio flujo de la issue.

Excluido:

- Filtros o ordenación por fecha límite.
- Notificaciones o alertas de vencimiento.
- Hora, zona horaria o recordatorios.
- Cambios en rutas (`actions/app.go`): no son necesarios.
- El cambio local no confirmado en `config/adw.yml`.
- Los `plans/` no rastreados del directorio de trabajo, y los archivos no rastreados `.txt`, `.x`, `migrations/schema.sql` y `.claude/skills/ship/`.

## Diseño por capas

- **Persistencia y migraciones**: columna `due_date date NULL` en `tasks`. Ya existe. Migración reversible. No se necesita índice.
- **Modelo** (`models/task.go`): campo `DueDate *time.Time`. Ya existe. Se mantiene `Validate` sin fecha porque el formato se valida al convertir el texto del formulario.
- **Acción y ruta** (`actions/tasks.go`): `parseDueDate` centraliza la conversión; Create/Update añaden el error a `verrs`. Patrón igual al del título (`verrs.Add`, 422 y re-render con `dueDateInput`). Ya existe.
- **Renderizado** (`templates/tasks/*.plush.html`): input de fecha, etiqueta en la lista, error accesible. Ya existe.
- **CSS** (`public/assets/app.css`): `.task-due` y el ancho del input de fecha. Ya existe.
- **JavaScript**: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz` y `.down.fizz`: añaden y quitan la columna.
- `models/task.go`: struct `Task` con `DueDate`.
- `models/task_test.go`: tests del modelo incluidos en la rama.
- `actions/tasks.go`: `parseDueDate`, handlers de tareas.
- `actions/tasks_test.go`: tests de integración de los handlers (suite con `as.authed`, CSRF incluido).
- `actions/actions_test.go`: pruebas adicionales de la suite (rama).
- `templates/tasks/index.plush.html`: lista y alta.
- `templates/tasks/edit.plush.html`: edición.
- `public/assets/app.css`: estilos de la etiqueta y del input de fecha.
- `database.yml`: bloque `test` con `database: todo_test`; los tests necesitan `GO_ENV=test`.

## Implementación paso a paso

La implementación ya está en la rama. Pasos para cerrar:

1. Confirmar que `git status -sb` muestra la rama `feat/issue-3-add-task-due-date` alineada con `origin` y que `config/adw.yml` sigue fuera del alcance.
2. Revisar cada criterio CA1–CA8 contra el código y los tests listados. Si uno no tiene test, añadirlo siguiendo el patrón de `Test_TasksCreate_*` / `Test_TasksUpdate_*` (suite `ActionSuite`, `as.authed`, `Post`/`Put` con `url.Values`).
3. Si un criterio no se cumple, corregir solo ese punto en el archivo afectado, sin tocar la arquitectura.
4. Ejecutar las validaciones de la sección siguiente con `GO_ENV=test`.
5. No hacer merge ni `--force`.

## Tests y validación

Tests que demuestran el comportamiento (ya en la rama):

- Crear con fecha (CA1): `Test_TasksCreate_WithDueDate`.
- Crear con fecha inválida (CA3): `Test_TasksCreate_InvalidDueDate`.
- Título inválido conserva fecha (CA4): `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`.
- Editar fija, cambia y borra (CA5): `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Precarga en edición (CA6): `Test_TasksEdit_PrefillsDueDate`.
- Toggle conserva fecha (CA7): `Test_TasksToggle_KeepsDueDate`.
- Lista muestra fecha (CA1): `Test_TasksIndex_ShowsDueDate`.
- Modelo: `models/task_test.go` (cambios de la rama).
- Migración (CA8): verificar que `up` y `down` se aplican en `todo_test`; no hay test automático específico.

Validaciones:

- `gofmt -l .`: en este Windows, los archivos del árbol de trabajo tienen finales de línea CRLF (`git ls-files --eol` muestra `i/lf w/crlf`) y gofmt los lista. Los blobs en git son LF y limpios; el listado es ruido de entorno. Para validar el contenido, ejecutar `gofmt -l` sobre los archivos tocados y comparar con el mismo comando en `main` (`git show main:<ruta> > <temporal>`), o convertir a LF solo en una copia temporal.
- `go vet ./...`: verificado en esta sesión, sin salida (OK).
- `go build ./...`: verificado en esta sesión (OK).
- `GO_ENV=test go test -p 1 ./...`: pendiente. Requiere PostgreSQL accesible con la base `todo_test` y migraciones aplicadas. En esta sesión no se encontró `psql` en el PATH; no se ha confirmado el servidor.

## Riesgos y recuperación

- **Ejecución de tests no confirmada**: si PostgreSQL no está disponible, los tests de acción no pueden ejecutarse. No abrir PR con esa validación sin reportarlo como pendiente.
- **Ejecución sin `GO_ENV=test`**: según la memoria del proyecto, la app escribe en `development`. Exportar `GO_ENV=test` antes de cualquier test.
- **Zona horaria**: la fecha se guarda como medianoche UTC; el formato `Format("2006-01-02")` evita desplazamientos en la lista y en el formulario. Si se cambia el formato, revisar este punto.
- **Migración**: `date` nullable, reversible con `drop_column`. Si la base ya tiene la columna con otro tipo, detenerse y revisar `migrations/schema.sql` (no rastreado).
- **Alcance de edición**: la edición (I2) no está explícita en la issue. Si el responsable quiere solo alta y visualización, retirar los cambios de `templates/tasks/edit.plush.html` y de `TasksEdit`/`TasksUpdate` en un cambio aparte; no hacerlo sin confirmación.
- **Recuperación**: la rama es reversible con commits nuevos; no reescribir historia.

## Siguiente paso

Ejecutar `/implement plans/issue-3-due-date-dia-estado-real-feature.md` para cerrar la verificación, correr la suite con `GO_ENV=test` y preparar el PR.
