# Feature: fecha límite opcional por tarea, solo día (issue #3)

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título «añadir fechaimie alas task» (leído con `gh issue view 3 --json`). Cuerpo literal: «vale un dia sin hora».

Requisitos explícitos:

- Las tareas pueden tener una fecha límite.
- La fecha límite es solo día, sin hora.

Criterios de aceptación observables:

1. Al crear una tarea se puede indicar una fecha opcional con un campo `type="date"` junto al título. Si se deja vacía, la tarea se crea sin fecha.
2. Al crear o editar una tarea con fecha válida, la fecha se guarda como día (`YYYY-MM-DD`, hora 00:00) y la lista muestra «Fecha límite: DD/MM/AAAA» bajo el título.
3. Al editar una tarea se puede cambiar o borrar la fecha límite. Borrarla deja `due_date` a NULL.
4. Una fecha no válida (p. ej. `no-es-fecha`) devuelve 422 con el mensaje «La fecha no es válida.» y conserva el valor introducido en el campo.
5. Un título vacío o inválido devuelve 422 y conserva la fecha introducida.
6. Completar o reabrir una tarea no altera su fecha límite.
7. Las tareas existentes (sin columna previa) siguen funcionando: la columna es nullable.

Inferencias (no explícitas en la issue, tomadas del código existente y de la forma en que está redactado el encargo):

- El nombre del campo en el formulario y en la base de datos es `due_date`.
- Se usa el input nativo `type="date"`; no se añade hora ni zona horaria.
- La fecha no tiene validación de rango (no se exige que sea futura).

## Comportamiento actual y propuesta

Estado de la rama `feat/issue-3-add-task-due-date` (ya comprometido, por encima de `main`): la feature está implementada de extremo a extremo en commits previos. Este plan no reescribe esa implementación; define qué queda por verificar y por cerrar antes de abrir el PR.

Implementado en la rama (verificado leyendo el diff `main...HEAD`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `down` (`drop_column`).
- `models/task.go`: campo `DueDate *time.Time` con tags `db:"due_date"`, `form:"due_date"`, `json:"due_date"`. El `Validate` solo valida el título.
- `actions/tasks.go`:
  - `parseDueDate(raw string) (*time.Time, error)`: cadena vacía → `nil`; si no, `time.Parse("2006-01-02", raw)`.
  - `TasksCreate` y `TasksUpdate` leen `title` y `due_date` con un struct de entrada, llaman a `task.Validate(tx)` y añaden el error `due_date` si el parseo falla. En error renderizan de nuevo con `dueDateInput` (valor crudo) para conservarlo. Al guardar usan `tx.Create` / `tx.Update`.
  - `TasksIndex` y `TasksEdit` fijan `dueDateInput` vacío.
- `templates/tasks/index.plush.html`: input `type="date"` `name="due_date"` en el formulario de creación; etiqueta «Fecha límite: …» en cada tarea; error `due_date` debajo del formulario.
- `templates/tasks/edit.plush.html`: input `type="date"` con etiqueta visible «Fecha límite», texto de ayuda «Opcional. Solo indica el día, sin hora.» y error asociado con `aria-describedby`.
- `public/assets/app.css`: clases `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Tests: `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`) y `actions/tasks_test.go` (creación con/sin fecha, fecha inválida, conservación de la fecha tras errores de título, mostrar fecha en índice, editar/borrar fecha, prerrelleno en edición, conservar fecha al completar). `actions/actions_test.go` añade el helper `authed` con token CSRF.

Pendiente o a revisar:

- Comprobar que todo compila, pasa `gofmt`, `go vet` y la suite completa con la base de prueba preparada (ver «Tests y validación»). Hasta ahora no hay evidencia de ejecución en esta sesión.
- `models/task.go` mantiene `form:"due_date"` en el modelo, aunque las acciones leen la fecha desde un struct de entrada. Es redundante pero inofensivo; no cambiarlo salvo que los tests lo exijan.

## Alcance

Incluido:

- Verificar la implementación ya presente en la rama (migración, modelo, acciones, plantillas, CSS, tests).
- Cerrar cualquier fallo que aparezca al ejecutar las validaciones, dentro de los archivos de esta feature.
- Preparar el PR desde esta rama contra `main`.

Excluido:

- Hora, zona horaria, recordatorios, filtros o orden por fecha límite.
- Validación de fecha futura o de rango.
- Cambios en `config/adw.yml`, en el directorio `plans/` de otros intentos, en `.claude/skills/ship/` y en archivos sin seguimiento que no sean de este cambio (`.txt`, `.x`, `migrations/schema.sql`).
- Los tests de `adws/` que fallan en Windows, ajenos a esta feature.

## Diseño por capas

- **Persistencia**: columna `due_date date NULL` en `tasks`. Migración ya creada; su `down` la elimina. No se usa `timestamp`, porque la issue pide solo día.
- **Modelo** (`models/task.go`): `*time.Time` para que «sin fecha» sea `NULL` y no una fecha cero.
- **Acción y ruta** (`actions/tasks.go`, `app.go`): las rutas existentes `POST /tasks` y `PUT /tasks/{task_id}` se reutilizan. La validación del título sigue en `Task.Validate`; la de la fecha vive en el controlador a través de `parseDueDate`, porque el error de parseo es de formato del formulario y no del modelo.
- **Renderizado** (`templates/tasks/*.plush.html`): input nativo, mensaje de error asociado y etiqueta de fecha en la lista.
- **CSS** (`public/assets/app.css`): solo los estilos mínimos de la etiqueta y del input de fecha. Sin JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añadir y quitar la columna `due_date` |
| `models/task.go` | Campo `DueDate` y validación del título |
| `models/task_test.go` | Persistencia de fecha opcional y solo día |
| `actions/tasks.go` | `parseDueDate`, creación y actualización con fecha, re-render con error |
| `actions/tasks_test.go` | Comportamiento HTTP de la feature |
| `actions/actions_test.go` | Helper `authed` con token CSRF para peticiones que cambian estado |
| `templates/tasks/index.plush.html` | Campo de fecha en el alta y etiqueta en la lista |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición y error asociado |
| `public/assets/app.css` | Estilos de `.task-due` y del input de fecha |

## Implementación paso a paso

No hay implementación nueva prevista. Pasos:

1. Confirmar que la rama está limpia en lo relativo a este cambio: `git status --short --branch`. Mantener fuera del PR `config/adw.yml` y los archivos sin seguimiento que no sean del cambio.
2. Ejecutar las validaciones de «Tests y validación». Si alguna falla por causa de esta feature, corregir el mínimo necesario dentro de los archivos listados y repetir.
3. Si todo pasa, no añadir tests nuevos salvo que falte un caso de la lista de criterios de aceptación. Los casos 1–7 ya tienen cobertura en `actions/tasks_test.go` o `models/task_test.go` según lo leído en el diff; confirmar esto al ejecutar.
4. Hacer stage solo de los archivos del cambio con una lista explícita, sin `git add .` ni `git add -A`.

## Tests y validación

Condición previa, obligatoria para la suite de acciones: la base de datos de prueba debe estar disponible (PostgreSQL con `todo_test` y migraciones aplicadas). Los tests de `actions/` deben ejecutarse con `GO_ENV=test`; sin esa variable la app escribe en `todo_development` y la suite lee `todo_test`, y los fallos en cascada no son reales.

Comandos, desde la raíz del repositorio:

```
gofmt -l .
go vet ./...
GO_ENV=test go test -p 1 -count=1 ./models/ ./actions/
go build ./...
```

Criterio de éxito: `gofmt -l .` sin salida; `go vet`, `go test` y `go build` sin error. Para la suite completa, `GO_ENV=test go test -p 1 -count=1 ./...` debe pasar salvo los fallos ya conocidos de `adws/` en Windows, que deben reportarse como preexistentes y no como fallos de esta feature (comprobar que también fallan en `main`).

Si `actions/` falla por 404, 422 inesperado o `sql: no rows` pese a `GO_ENV=test`, revisar primero que la migración `20261022120000` está aplicada en `todo_test`.

## Riesgos y recuperación

- **Zona horaria**: `time.Parse` sin zona devuelve UTC medianoche. Al mostrar con `Format` no hay desplazamiento, pero si se cambia el tipo a `timestamp` la fecha podría desplazarse un día. Mantener `date`.
- **Migración ya aplicada en otros entornos**: si la migración ya corrió, no editar su contenido; crear una nueva si hace falta un cambio.
- **Valor crudo en re-render**: `dueDateInput` se escribe en un atributo HTML con escapado de Plush; el test `Test_TasksCreate_InvalidDueDate` comprueba que aparece el valor introducido.
- **Recuperación**: si la validación falla y el cambio no es recuperable en el plan, no abrir PR; revertir solo los archivos modificados en este paso con `git checkout -- <archivo>` tras confirmar qué se va a perder.
