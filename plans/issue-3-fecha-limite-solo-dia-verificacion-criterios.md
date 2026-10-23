# Feature: Fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Encargo (issue #3, "añadir fechaimie alas task", etiqueta `adw`): añadir una fecha límite a cada tarea. Comentario de la issue: "vale un día sin hora".

Requisitos explícitos:

1. Las tareas pueden tener una fecha límite.
2. La fecha límite es solo de día: no tiene hora.

Inferencias (no escritas en la issue, asumidas por coherencia con el resto de la app):

3. La fecha límite es opcional: una tarea puede no tener fecha.
4. Se puede indicar al crear y cambiar al editar.
5. Se muestra en la lista de tareas.
6. Una fecha mal formada no se guarda y devuelve error de validación (422).
7. Completar o reabrir una tarea no altera su fecha límite.

Criterios de aceptación:

- CA1: crear tarea sin fecha deja `due_date` nulo.
- CA2: crear tarea con fecha `YYYY-MM-DD` la persiste como día, sin componente horario.
- CA3: editar una tarea permite establecer, cambiar y borrar la fecha.
- CA4: una fecha inválida devuelve 422 con mensaje y no modifica la base de datos.
- CA5: si falla la validación del título, el formulario conserva la fecha enviada.
- CA6: la lista muestra la fecha en formato `dd/mm/aaaa` solo cuando existe.
- CA7: el formulario de edición precarga la fecha existente en `YYYY-MM-DD`.
- CA8: alternar completada no cambia la fecha límite.

## Comportamiento actual y propuesta

Estado verificado en la rama `feat/issue-3-add-task-due-date` (leído del código, no ejecutado contra base de datos):

- Commit `d13ec2c` `feat(tasks): añadir fecha límite opcional (solo día) a las tareas` (cita `Resuelve #3`): migración, modelo, acciones, plantillas y tests.
- Commit `69e8826` `fix(tasks): conservar la fecha límite al fallar la validación del alta`: corrige CA5 y añade test de regresión.
- Ambos commits ya están en la rama; el árbol de trabajo no tiene cambios de aplicación pendientes para esta feature.

Cobertura por criterio (estado del código):

| Criterio | Implementado en | Test que lo cubre |
|---|---|---|
| CA1 | `models/task.go` (`DueDate *time.Time`), `actions/tasks.go` (`parseDueDate` devuelve `nil` si vacío) | `models/task_test.go` `Test_Task_DueDate_Optional`, `actions/tasks_test.go` `Test_TasksCreate_Valid` |
| CA2 | migración `add_column("tasks", "due_date", "date", {"null": true})` | `models/task_test.go` `Test_Task_DueDate_DayOnly_Persists` (hora 0) |
| CA3 | `TasksUpdate` asigna `task.DueDate` desde el formulario | `Test_TasksCreate_WithDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate` |
| CA4 | `parseDueDate` + `verrs.Add("due_date", dueDateInvalidMessage)` antes de guardar | `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate` |
| CA5 | `TasksCreate` vuelve a renderizar con `task` que conserva `DueDate` | `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate` |
| CA6 | `templates/tasks/index.plush.html` `t.DueDate.Format("02/01/2006")` | `Test_TasksIndex_ShowsDueDate` |
| CA7 | `templates/tasks/edit.plush.html` `value` con `Format("2006-01-02")` | `Test_TasksEdit_PrefillsDueDate` |
| CA8 | `TasksToggle` solo invierte `Completed` | `Test_TasksToggle_KeepsDueDate` |

Propuesta: no escribir código nuevo. Verificar que los tests y la validación pasan en un entorno con PostgreSQL y, si es así, abrir el PR de la rama. Cualquier fallo se trata como bloqueo y no se corrige alcance adicional.

## Alcance

Incluido:

- Verificar CA1–CA8 con los tests existentes.
- Ejecutar la validación de CI (formato, vet, migraciones de test, tests, build).
- Preparar el PR de la rama con solo los archivos de la feature y de este plan.

Excluido:

- Hora de vencimiento, zonas horarias o recordatorios.
- Ordenar, filtrar o resaltar tareas vencidas por fecha.
- Cambios en `config/adw.yml`, `migrations/schema.sql` y `.claude/skills/ship/`: son cambios sin confirmar previos a esta tarea y no forman parte de la feature.
- Los planes `plans/issue-3-*.md` y `plans/verificar-fecha-limite-tareas.md` ya existentes: son registros de trabajo anteriores; no se modifican.

## Diseño por capas

Afectadas: persistencia, modelo, acciones y plantillas. JavaScript y CSS no cambian.

- Persistencia: `migrations/20261022120000_add_due_date_to_tasks.up.fizz` añade `due_date date null`. `.down.fizz` la elimina. Es aditiva y compatible con filas existentes (quedan con `NULL`).
- Modelo: `models/task.go` `DueDate *time.Time` con etiqueta `form:"due_date"`. El tipo puntero representa "sin fecha".
- Acciones: `actions/tasks.go` concentra el parseo en `parseDueDate` (formato `2006-01-02`) y reutiliza `dueDateInvalidMessage`. Create y Update usan el mismo patrón de validación; no hay una arquitectura paralela.
- Plantillas: `templates/tasks/index.plush.html` (campo en el alta y fecha mostrada) y `templates/tasks/edit.plush.html` (campo de edición con ayuda "Opcional. Solo indica el día, sin hora.").

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date` como `date` nullable |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna |
| `models/task.go` | Campo `DueDate` y validación del título |
| `models/task_test.go` | Tests de persistencia de la fecha |
| `actions/tasks.go` | `parseDueDate`, Create, Update, Toggle |
| `actions/tasks_test.go` | Tests HTTP de alta, edición, errores y visualización |
| `templates/tasks/index.plush.html` | Campo de fecha en el alta y fecha en la lista |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición |
| `.github/workflows/ci.yaml` | Secuencia de validación que replicar en local |

## Implementación paso a paso

1. Confirmar que la rama contiene `d13ec2c` y `69e8826` (`git log --oneline -3`).
2. Confirmar que no hay cambios de aplicación sin confirmar: solo `config/adw.yml`, `migrations/schema.sql` y `.claude/skills/ship/` aparecen fuera del commit.
3. Ejecutar la validación de la sección siguiente.
4. Si todo pasa: no hay cambios de código. Pasar al PR con los archivos del plan.
5. Si algún test falla: detenerse, registrar la salida y no abrir el PR.

## Tests y validación

Tests que demuestran el comportamiento: los listados en la tabla de cobertura. Son suficientes para CA1–CA8; no se añaden tests nuevos salvo que la validación revele un hueco real.

Condiciones necesarias para ejecutar los tests de aplicación y modelo:

- PostgreSQL 17 accesible con usuario `postgres`, contraseña `postgres` y base `todo_test` (como en `ci.yaml`).
- Migraciones de test aplicadas: `soda migrate -e test` (versión `v6.4.1` del CLI, según CI).
- Variable de entorno `GO_ENV=test`.

Comandos (orden de CI):

```
gofmt -l .
go vet ./...
soda migrate -e test
go test -p 1 ./...
go build ./...
```

Notas sobre el entorno de esta sesión:

- `go` 1.27.1 está disponible; `psql` y `pg_isready` no están en el PATH, así que los tests con base de datos no se han ejecutado aquí.
- `gofmt -l .` lista casi todos los `.go`, incluidos los que no toca esta feature, porque el árbol de trabajo tiene terminaciones CRLF (`git ls-files --eol` muestra `i/lf w/crlf`) y en el índice son LF. Es un artefacto del checkout de Windows, no un fallo de formato; en CI (Linux) el resultado es el válido. Ignorar esos archivos solo si se confirma que el único motivo es el fin de línea (por ejemplo, comparando con `git show HEAD:<ruta> | gofmt -l` en un archivo temporal).
- Los directorios `.claude/worktrees/*` aparecen en `gofmt -l` y en el recorrido de paquetes; no forman parte del repositorio principal y no deben incluirse en el PR.

## Riesgos y recuperación

- Falsa sensación de cierre: el código ya existe y los commits de documentación anteriores afirman que los tests pasan. Este plan exige ejecutar la validación en el estado actual antes de dar el trabajo por verificado.
- Entorno: sin PostgreSQL no hay verificación de tests de base de datos; el resultado de la validación debe declararse como no ejecutado en ese caso.
- Fin de línea CRLF: si se confirman cambios de aplicación en Windows, el diff puede contener cambios de fin de línea no deseados. Revisar con `git diff --stat` antes de stage.
- Cambios sin confirmar ajenos (`config/adw.yml`, `migrations/schema.sql`, `.claude/skills/ship/`): no entran en el PR. `config/adw.yml` cambia el prompt del trigger; no debe mezclarse con esta feature.
- Migración: si hubiera que revertirla, `down.fizz` elimina la columna y pierde las fechas guardadas. Es reversible solo sin datos que conservar.
- Recuperación: si la validación falla, no hay PR que deshacer; corregir en la rama con un commit nuevo, sin `--force`.

Siguiente paso sugerido: `/implement plans/issue-3-fecha-limite-solo-dia-verificacion-criterios.md`.
