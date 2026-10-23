# Feature: cierre de la fecha límite (solo día) de las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". Lectura adoptada: **cada tarea puede tener una fecha límite opcional, con precisión de día (`YYYY-MM-DD`) y sin hora**.

Requisitos explícitos (del cuerpo de la issue):
- La fecha es solo de día, sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional.
- Se indica al crear y al editar.
- Se muestra en la lista (`DD/MM/YYYY`) y se precarga al editar.

Criterios de aceptación:
1. Crear una tarea sin fecha funciona como antes.
2. Crear una tarea con `2026-10-22` la guarda como día sin hora y la lista muestra `22/10/2026`.
3. Editar permite cambiar o borrar la fecha sin alterar título ni `Completed`.
4. Una fecha mal formada no se guarda: 422 con mensaje en `due_date`, nunca 500.
5. La migración añade una columna `date` nullable; las tareas existentes quedan sin fecha.

## Comportamiento actual y propuesta

La funcionalidad **ya está implementada** en la rama `feat/issue-3-add-task-due-date` (commits `d13ec2c` y `7913cf0`, pendientes de PR contra `main`). Comprobado leyendo el código:

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`; el `.down.fizz` hace `drop_column`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` (vacío → `nil`; si no, `time.Parse("2006-01-02")`). `TasksCreate` y `TasksUpdate` leen `due_date` como string, añaden `dueDateInvalidMessage` a `due_date` y responden 422 si falla. `TasksUpdate` asigna `task.DueDate` en cada PUT.
- `templates/tasks/index.plush.html`: input `type="date"` en el alta; muestra `Fecha límite: DD/MM/YYYY` si `DueDate != nil`.
- `templates/tasks/edit.plush.html`: input `type="date"` precargado en `YYYY-MM-DD`, con ayuda y error.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]`.
- Tests existentes: `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`) y `actions/tasks_test.go` (`Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`).

Lo que **no** está hecho:

- Ninguno de estos tests ni comandos de validación se ha ejecutado todavía en esta rama.
- **Gap de UX (R8)**: tras un 422 en el alta, `index.plush.html` renderiza el input de fecha vacío porque no pasa `task.DueDate` a `value`. El título sí se conserva. No rompe los criterios 1 a 5, pero es un fallo visible del flujo de alta con fecha.
- **Decisión pendiente (R6)**: un PUT sin `due_date` borra la fecha. Lo verifica `TasksUpdate`; el formulario de edición siempre envía el campo, así que solo afecta a clientes que omitan el campo.

Propuesta: no rediseñar. (a) Ejecutar la validación completa. (b) Corregir R8 con el cambio mínimo en la plantilla y un test de regresión. (c) Mantener el comportamiento de R6 (reemplazo completo) salvo que el usuario decida otra cosa en el checkpoint.

## Alcance

Incluido:
- Ejecutar la validación de la rama (gofmt, vet, migración de test, tests, build) y registrar el resultado.
- Comprobar la reversión de la migración (criterio 5).
- Corregir R8: en `templates/tasks/index.plush.html`, que el input `due_date` del alta muestre el valor enviado tras un 422, con el mismo formato que la edición (`YYYY-MM-DD`).
- Añadir un test de regresión para R8 en `actions/tasks_test.go`, siguiendo el patrón de `Test_TasksCreate_InvalidDueDate`.
- Corregir solo lo que falle en la validación, dentro de los criterios 1 a 5.

Excluido:
- Hora, zona horaria, recordatorios o notificaciones.
- Ordenar o filtrar por fecha; resaltar tareas vencidas.
- Cambiar la semántica de PUT sin `due_date` (R6). Es una decisión del usuario.
- Cambiar `TasksToggle`, `TasksDestroy`, rutas, `fixtures/` o el seed de migraciones.
- Tocar `config/adw.yml`, `plans/add-task-due-date.md`, `migrations/schema.sql`, `.claude/skills/ship/` ni los planes `plans/issue-3-verificar-fecha-limite.md`, `plans/verificar-fecha-limite-tareas.md`, `plans/issue-3-fecha-limite-verificacion.md`, `plans/issue-3-add-task-due-date.md`. Son ajenos al cambio o ya existen; no entran en el PR salvo este plan.

## Diseño por capas

Capas afectadas:

- **Renderizado** (`templates/tasks/index.plush.html`): pasar `value` al input `due_date` del alta, reflejando `task.DueDate` cuando no es nil. Patrón igual al de `edit.plush.html`.
- **Persistencia, modelo, acción, CSS**: ya implementados; sin cambios.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` | Columna nullable y su reversión |
| `models/task.go` | Campo `DueDate` |
| `actions/tasks.go` | `parseDueDate`, alta y edición con `due_date` |
| `templates/tasks/index.plush.html` | Alta con fecha (a corregir: R8) y lista |
| `templates/tasks/edit.plush.html` | Edición con fecha precargada |
| `actions/tasks_test.go` | Tests de acción; aquí se añade el de R8 |
| `models/task_test.go` | Tests de modelo |
| `.github/workflows/ci.yaml` | Solo lectura: orden de validación del CI |

## Implementación paso a paso

1. Confirmar que la rama contiene `d13ec2c` y `7913cf0` y que el árbol no tiene cambios en los archivos de la tabla anterior.
2. Preparar la base de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks` sin error.
3. Ejecutar la validación inicial (sección siguiente) **antes de tocar código**, para tener línea base.
4. Escribir primero el test de R8 en `actions/tasks_test.go`: un POST `/tasks` con título vacío y `due_date` válido debe responder 422 y el HTML debe contener el valor `value="2026-10-22"` en el input `due_date`. Debe fallar con el código actual.
5. Corregir `templates/tasks/index.plush.html`: en el input `due_date` añadir `value="<%= if (task.DueDate != nil) { %><%= task.DueDate.Format("2006-01-02") %><% } %>"`, igual que en `edit.plush.html`. Si el 422 viene de un `due_date` inválido, el valor no existe y el input queda vacío (correcto).
6. Repetir la validación completa; el test de R8 debe pasar y ningún otro debe cambiar.
7. Comprobar la reversión (criterio 5): `soda migrate down -e test` y después `soda migrate -e test`.
8. No cambiar `TasksUpdate` por R6; la decisión se registra en el reporte.

## Tests y validación

Comandos, en el orden del CI:

- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → la migración se aplica.
- `go test -p 1 ./...` → todo en verde, incluido el test nuevo de R8.
- `go build ./...` → compila.
- Reversión: `soda migrate down -e test` seguido de `soda migrate -e test`.

Cobertura por criterio:

- Criterio 1: `Test_TasksCreate_Valid`, `Test_Task_Create_Valid`.
- Criterio 2: `Test_TasksCreate_WithDueDate`, `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3: `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`.
- Criterio 4: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5: migración y reversión (paso 7).
- R8: test nuevo (paso 4).

Condiciones necesarias:

- Go instalado y la versión de `go.mod`.
- PostgreSQL 17 accesible con la configuración de `database.yml` para `test` (en CI: usuario y contraseña `postgres`, base `todo_test`). En la máquina de trabajo el puerto 5432 responde, pero no hay `psql` ni `pg_isready`; la conexión **no está verificada**. Si no hay base, reportar que los tests no se pudieron ejecutar; no declarar la validación superada.
- `soda` instalado (`go install github.com/gobuffalo/pop/v6/soda@v6.4.1`, como en CI; en esta máquina está en `~/go/bin/soda`).

## Riesgos y recuperación

- **R6, PUT sin `due_date` borra la fecha.** Decisión pendiente. Recomendación: mantener el reemplazo completo, porque el formulario de edición siempre envía el campo. Si el usuario prefiere conservar la fecha cuando el campo no llega, el cambio es condicionar la asignación a la presencia de la clave en el input; eso requiere un test nuevo y no está en este plan.
- **R8, corrección en plantilla.** Riesgo bajo: solo cambia un atributo `value`. Recuperación: revertir la línea de `index.plush.html` y el test.
- **R2, zona horaria.** `time.Parse` sin zona devuelve medianoche UTC; la columna es `date`, así que no hay desplazamiento de día.
- **R3, migración en entornos existentes.** `ADD COLUMN ... NULL` no reescribe filas en PostgreSQL. Recuperación: `soda migrate down -e <entorno>`.
- **Entorno sin base.** Si PostgreSQL no responde, los tests de acción y modelo no se ejecutan; reportarlo tal cual.
- **Rama con cambios ajenos.** El árbol tiene cambios sin confirmar que no son de esta feature; no se incluyen en ningún commit ni en el PR.

## Resultado de la validación (2026-10-23)

- R8 ya estaba corregido en la rama (`templates/tasks/index.plush.html` pasa `value` al input `due_date`, commit `69e8826`) y cubierto por `Test_TasksCreate_InvalidTitleKeepsDueDate`. No hizo falta cambiar código en esta ejecución.
- `go vet ./...`, `go build ./...`: sin errores.
- `gofmt` sobre el contenido commiteado (`git show HEAD:<archivo> | gofmt -l`): limpio. `gofmt -l .` sobre el disco marca todos los `.go` porque el checkout de Windows tiene CRLF (`core.autocrlf=true`); no es un problema del código.
- `soda migrate -e test`: migraciones al día.
- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/`: OK, incluidas todas las pruebas de fecha.
- `go test -p 1 ./...`: falla solo `todo/adws` (`TestTrigger*`). Ese paquete no cambia en la rama; los tests usan un `gh` falso sin extensión que en Windows no se resuelve y se ejecuta el `gh` real contra `example/todo`. Pendiente de confirmar en un entorno Linux o con el `gh` falso como `.cmd`.
