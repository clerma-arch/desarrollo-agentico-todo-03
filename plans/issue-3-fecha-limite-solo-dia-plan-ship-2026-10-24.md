# Feature: fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Encargo (issue #3, "añadir fecha límite a las tasks"): cada tarea puede tener una fecha límite. El cuerpo de la issue dice "vale un dia sin hora": la fecha es solo día, sin hora.

Criterios de aceptación:

1. Crear una tarea sin fecha sigue funcionando; la fecha es opcional.
2. Crear una tarea con fecha válida (`YYYY-MM-DD`) guarda ese día, sin hora.
3. Editar una tarea permite fijar, cambiar o borrar la fecha.
4. Una fecha con formato inválido no se guarda y muestra un error en el campo `due_date`.
5. Si falla la validación del título, la fecha introducida se conserva en el formulario.
6. La lista muestra la fecha como `DD/MM/AAAA`, sin hora.
7. Completar o reabrir una tarea conserva su fecha límite.

Requisitos explícitos (de la issue): fecha límite por tarea, solo día.

Inferencias (no explícitas en la issue): que la fecha sea opcional, que se valide en servidor con formato `YYYY-MM-DD`, y que la edición pueda borrarla. Son coherentes con el código ya existente en la rama.

## Comportamiento actual y propuesta

Estado de la rama `feat/issue-3-add-task-due-date` frente a `main` (commits ya presentes, sin subir a un PR):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `down`.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` solo valida el título.
- `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en `*time.Time` (vacío = nil). `TasksCreate` y `TasksUpdate` leen `due_date`, validan con `task.Validate(tx)` y añaden el error `due_date` si el parseo falla. `TasksEdit` y el alta fallida conservan `dueDateInput`.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de alta y muestra `Fecha límite: DD/MM/AAAA` en cada tarea.
- `templates/tasks/edit.plush.html`: input `type="date"` con ayuda y error accesible (`aria-describedby`, `role="alert"`).
- `public/assets/app.css`: estilo `.task-due` y ancho del input de fecha en `.new-task-row`.
- Tests en `actions/tasks_test.go` (crear con/sin fecha, fecha inválida, conservar fecha en error de título, actualizar/borrar, completar, mostrar y prellenar) y `models/task_test.go`.

Propuesta: no hay cambios de diseño pendientes. El paso de implementación consiste en verificar que la rama cumple los criterios, ejecutar las validaciones y, si alguna falla, corregir solo lo necesario dentro de este alcance.

## Alcance

Incluido:

- Persistencia (migración `date` nullable), modelo, acciones de alta/edición, plantillas y CSS mínimo ya presentes en la rama.
- Tests que demuestran los criterios de aceptación.

Excluido:

- Hora de vencimiento, recordatorios, notificaciones, ordenación o filtro por fecha.
- Cambios en `config/adw.yml`, `.claude/skills/ship/`, `.txt`, `.x` y `migrations/schema.sql` (vacío y sin seguimiento). No forman parte del cambio y no deben entrar en el PR.
- Planes antiguos bajo `plans/` que no son de este cambio (se commitean solo los de esta feature si hace falta).

## Diseño por capas

- **Persistencia**: columna `due_date date NULL` en `tasks`. Migración ya creada con `up` y `down`.
- **Modelo**: `*time.Time` para representar ausencia de fecha. `Validate` mantiene solo la regla del título; el formato se valida en la acción porque el modelo recibe ya un `time.Time`.
- **Acción y ruta**: `parseDueDate` centraliza el formato `2006-01-02`. Las rutas no cambian (`/tasks`, `/tasks/{id}`).
- **Renderizado**: Plush con `Format("02/01/2006")` para la lista y `Format("2006-01-02")` para el valor del input.
- **CSS**: solo dos reglas; sin JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y retira la columna `due_date`. |
| `models/task.go` | Campo `DueDate` y validación del título. |
| `models/task_test.go` | Pruebas del modelo. |
| `actions/tasks.go` | `parseDueDate`, alta, edición y mensajes de error. |
| `actions/tasks_test.go` | Pruebas de integración de las acciones y de las plantillas. |
| `actions/actions_test.go` | Suite de tests de aplicación (`Test_ActionSuite`). |
| `templates/tasks/index.plush.html` | Input de fecha en el alta y fecha mostrada por tarea. |
| `templates/tasks/edit.plush.html` | Input de fecha en la edición, con error accesible. |
| `public/assets/app.css` | `.task-due` y ancho del input de fecha. |

## Implementación paso a paso

El cambio ya está en la rama. Pasos para llegar a un PR:

1. Confirmar que el árbol de trabajo tiene solo cambios de esta feature. `config/adw.yml` y los archivos sin seguimiento listados en "Alcance" quedan fuera.
2. Ejecutar las validaciones de "Tests y validación" con `GO_ENV=test`.
3. Si todo pasa, no hay cambios de código. Si alguna prueba falla, corregir únicamente el fallo, sin ampliar el alcance.
4. Preparar el PR desde `feat/issue-3-add-task-due-date` hacia `main`, con `git add` de la lista explícita de archivos del cambio y del plan.

## Tests y validación

Condiciones:

- Go 1.27 (instalado: `go1.27.1 windows/amd64`).
- PostgreSQL escuchando en `127.0.0.1:5432` (comprobado: el puerto está abierto).
- Base `todo_test` con la migración `20261022120000` aplicada.
- Variable `GO_ENV=test` (no está definida en la sesión actual). Sin ella, `actions/` escribe en `development`.

Comandos:

```
GO_ENV=test gofmt -l .
GO_ENV=test go vet ./...
GO_ENV=test go test -p 1 ./...
GO_ENV=test go build ./...
```

Tests que demuestran el comportamiento:

- `models/task_test.go`: validación del título y de la fecha como `*time.Time`.
- `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksToggle_KeepsDueDate`.

Estado: no se han ejecutado en esta etapa de planificación. Su resultado es una incertidumbre a cerrar en la implementación.

## Riesgos y recuperación

- **Zona horaria**: `time.Parse` produce medianoche UTC. Si la columna `date` o el formato de visualización cambian, el día mostrado podría desplazarse. Los tests de `Format` lo cubren con fechas UTC.
- **Migración**: si la base de test no tiene `due_date`, los tests fallarán con error de columna. Recuperación: aplicar la migración `up`; el `down` la retira.
- **Validación de título y fecha**: `TasksCreate` usa `Validate` seguido de `Create`, en lugar de `ValidateAndCreate`. Es equivalente porque `Validate` no devuelve error de base de datos, pero es un cambio de patrón a revisar en el PR.
- **Archivos sin seguimiento**: incluirlos por error contaminaría el PR. Usar solo rutas explícitas.
- **Revertir**: `git revert` de los commits de la feature, o `drop_column` con la migración `down`.

## Incertidumbres pendientes

- Si las pruebas actuales pasan (no ejecutadas aquí).
- Si la base `todo_test` tiene la migración aplicada.
- Si `GO_ENV=test` debe fijarse en el entorno de ejecución o en un script.
- Si los planes antiguos de `plans/` deben quedar fuera del PR (se propone excluirlos).

Siguiente paso: `/implement plans/issue-3-fecha-limite-solo-dia-plan-ship-2026-10-24.md`.
