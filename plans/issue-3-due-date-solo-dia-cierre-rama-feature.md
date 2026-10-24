# Feature: fecha límite opcional (solo día) en las tareas — cierre de la rama (issue #3)

Rama de trabajo: `feat/issue-3-add-task-due-date` (base `main`, remoto `origin`).
Estado de partida: la implementación ya está commiteada en la rama. Este plan verifica que cumple los criterios de la issue, fija lo que falta para cerrarla y no rediseña nada.

## Encargo y criterios de aceptación

Issue #3, "añadir fechaimie alas task", con cuerpo "vale un dia sin hora".

Requisitos explícitos (de la issue):

- R1. Cada tarea puede tener una fecha límite.
- R2. La fecha límite es opcional.
- R3. La fecha límite es solo día, sin hora.

Inferencias (no están en la issue; se asumen por coherencia con la app):

- I1. Formato de entrada `YYYY-MM-DD` (input `type="date"`); visualización `dd/mm/yyyy`.
- I2. Fecha inválida: se rechaza con `422`, mensaje en español y el valor enviado se conserva en el formulario.
- I3. Editar permite fijar y quitar la fecha.
- I4. Completar o descompletar una tarea no modifica su fecha.
- I5. La migración es reversible.

Criterios de aceptación verificables:

| ID | Criterio | Cubierto por (estado actual del código) |
|----|----------|------------------------------------------|
| CA1 | Alta con fecha guarda solo el día | `Test_TasksCreate_WithDueDate`, `models/task_test.go` `Test_Task_DueDate_DayOnly_Persists` |
| CA2 | Alta sin fecha deja `due_date` nulo | `Test_TasksCreate` (`as.Nil(task.DueDate)`), `models/task_test.go` `Test_Task_DueDate_Optional` |
| CA3 | Fecha inválida en alta: 422, mensaje y valor conservado | `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate` |
| CA4 | Fecha inválida en edición: 422, tarea sin cambios | `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate` |
| CA5 | Editar fija y quita la fecha | `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate` |
| CA6 | La lista muestra la fecha como `dd/mm/yyyy` | `Test_TasksIndex_ShowsDueDate` |
| CA7 | Completar conserva la fecha | `Test_TasksToggle_KeepsDueDate`, `0125661` |
| CA8 | Migración reversible | `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` |
| CA9 | Las escrituras de tests envían token CSRF | helper `csrfToken` en `actions/tasks_test.go`, commit `d222c26` |

Estos criterios están cubiertos por tests en el código, pero **no se han ejecutado en esta sesión**. La validación de la sección "Tests y validación" es la que los confirma.

## Comportamiento actual y propuesta

Comportamiento actual (leído del código de la rama):

- `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en `*time.Time` sin hora. Cadena vacía devuelve `nil`. `TasksCreate` y `TasksUpdate` validan el título y la fecha; si la fecha es inválida, añaden el error `due_date` y renderizan `422` con `dueDateInput` = texto enviado.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` solo valida el título.
- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. El down hace `drop_column`.
- `templates/tasks/index.plush.html` y `templates/tasks/edit.plush.html`: input `type="date"` `name="due_date"`, ayuda "Opcional. Solo indica el día, sin hora.", error `#due-date-error` con `role="alert"`, y `aria-invalid`/`aria-describedby` cuando hay error. La lista muestra `Fecha límite: dd/mm/yyyy`.
- `public/assets/app.css`: clase `.task-due`.

Propuesta: **no cambiar comportamiento**. Verificar, y solo si una prueba falla, corregir el mínimo en el archivo responsable.

## Alcance

Incluido:

- Ejecutar la validación de la rama y confirmar los criterios CA1–CA9.
- Corregir únicamente un fallo de test o de build que se reproduzca.

Excluido:

- Hora, zona horaria o recordatorios.
- Filtros o orden por fecha límite.
- Cambios en `config/adw.yml` (modificación previa sin commitear, no forma parte de la feature).
- Los planes sin commitear de `plans/` y los archivos sin seguimiento `.txt`, `.x`, `migrations/schema.sql`, `.claude/skills/ship/`. No entran en el PR.
- Nuevos tests fuera de los criterios CA1–CA9, salvo que un criterio no tenga cobertura real (ver paso 4).

## Diseño por capas

- **Persistencia y migraciones**: ya existe. Columna `due_date date NULL`. Sin cambios.
- **Modelo**: ya existe. `DueDate *time.Time`. Sin cambios.
- **Acción y ruta**: ya existen `TasksCreate`, `TasksUpdate`, `TasksEdit` y `TasksIndex`. Sin cambios salvo fallo de test.
- **Renderizado**: ya existe en las dos plantillas. Sin cambios.
- **JavaScript**: no aplica.
- **CSS**: `.task-due` existe. Sin cambios.

## Archivos relevantes

| Ruta | Responsabilidad |
|------|-----------------|
| `actions/tasks.go` | `parseDueDate`, acciones de alta, edición, lista y toggle |
| `actions/tasks_test.go` | Tests de acción de la feature y helper CSRF `csrfToken` |
| `actions/actions_test.go` | Setup de la suite de acciones (`todo_test`) |
| `models/task.go` | Campo `DueDate` y validación del título |
| `models/task_test.go` | Tests de persistencia de la fecha |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Columna `due_date` reversible |
| `templates/tasks/index.plush.html` | Input de fecha en alta, error y visualización en la lista |
| `templates/tasks/edit.plush.html` | Input de fecha en edición, ayuda y error |
| `public/assets/app.css` | Clase `.task-due` |
| `database.yml` | Base `todo_test` en `127.0.0.1:5432` |

## Implementación paso a paso

1. Establecer la línea base en la rama:
   - `git status --short --branch`: confirmar que la rama es `feat/issue-3-add-task-due-date`.
   - No añadir ni tocar nada de lo que aparece como no commiteado.
2. Comprobar formato y compilación sin tests:
   - `gofmt -l .`
   - `go vet ./...`
   - `go build ./...`
3. Ejecutar los tests afectados con el entorno de test. Son obligatorios `GO_ENV=test` (la suite de acciones usa `todo_test`; sin esa variable escribe en `development`) y PostgreSQL disponible en `127.0.0.1:5432`:
   - `GO_ENV=test go test -p 1 ./models/ ./actions/`
   - Si la base de test no tiene el esquema, aplicar las migraciones de test antes (`buffalo test` las prepara; no ejecutar migraciones contra `todo_development`).
4. Contrastar cada criterio CA1–CA9 con el test que lo cubre, leyendo el test, no solo su nombre. Si un criterio no tiene una aserción real, dejarlo registrado en el informe; no ampliar el alcance sin justificarlo.
5. Si algún test falla:
   - Reproducirlo y leer la salida completa.
   - Corregir solo el archivo responsable, con el cambio mínimo.
   - Repetir el paso 3.
   - Si el fallo exige cambiar comportamiento fuera de CA1–CA9, detenerse y explicar el bloqueo.
6. Ejecutar la validación completa:
   - `gofmt -l .` (debe salir vacío)
   - `go vet ./...`
   - `GO_ENV=test go test -p 1 ./...`
   - `go build ./...`
7. No hacer commit de cambios de código si no hubo ningún fallo. Si el paso 5 cambió algo, commitear solo esos archivos con Conventional Commits.

## Tests y validación

Comprobaciones pertinentes a esta feature:

- Formato: `gofmt -l .` sin salida.
- Análisis estático: `go vet ./...` sin errores.
- Modelo: `GO_ENV=test go test -p 1 ./models/` (persistencia del día, nulo, sin hora).
- Acciones: `GO_ENV=test go test -p 1 ./actions/` (alta, edición, lista, toggle, 422 y CSRF).
- Compilación: `go build ./...`.
- Suite completa: `GO_ENV=test go test -p 1 ./...` como validación final.

Condiciones necesarias:

- PostgreSQL escuchando en `127.0.0.1:5432` (comprobado en esta sesión: abierto).
- Base `todo_test` con migraciones aplicadas, incluida `20261022120000`.
- `GO_ENV=test` en el entorno de cada comando de test.

Criterio de paso: todos los comandos sin errores. Si la base no está disponible, el criterio no se cumple y se informa, sin abrir PR.

## Riesgos y recuperación

- **Zona horaria**: `time.Parse` sin ubicación produce UTC a medianoche. Si el driver de PostgreSQL devuelve el `date` con desplazamiento local, la fecha podría verse un día antes o después. Los tests de modelo comprueban `Hour() == 0` y el día; si fallan en un entorno con otra zona horaria, es el riesgo a revisar.
- **Estado de la base de test**: si `todo_test` no tiene la migración, los tests de acción fallarán de forma poco clara. Solución: aplicar migraciones de test, nunca de desarrollo.
- **Riesgo de alcance**: el PR de la feature ya existe (#5, abierto). Cualquier cambio nuevo debe ir a la misma rama para no duplicar PR.
- **Recuperación**: si hay que deshacer la migración, el down `drop_column("tasks", "due_date")` la revierte. Si hay que deshacer código, `git revert` de los commits concretos. No usar `--force`.
