# Feature: fecha límite de tareas, solo día (issue #3)

## Encargo y criterios de aceptación

Encargo literal de la issue #3 (`clerma-arch/desarrollo-agentico-todo-03`, abierta, sin comentarios, etiqueta `adw`):

- Título: "añadir fechaimie alas task" (sic; se refiere a "añadir fecha límite a las tareas").
- Cuerpo: "vale un dia sin hora".

Requisitos explícitos:

1. Las tareas pueden tener una fecha límite.
2. La fecha es solo día, sin hora ("vale un día sin hora").

Inferencias (no escritas en la issue, pero coherentes con el resto de la app):

3. La fecha es opcional: una tarea sin fecha sigue siendo válida.
4. El formulario de creación y el de edición aceptan la fecha en formato `YYYY-MM-DD` (control `input type="date"`).
5. La lista muestra la fecha como `dd/mm/aaaa` solo si existe.
6. Una fecha no válida se rechaza con un mensaje en español y el valor escrito se conserva al re-renderizar.
7. Completar o reabrir una tarea no altera su fecha límite.
8. Editar permite fijar, cambiar y borrar la fecha.

## Comportamiento actual y propuesta

**Estado de la rama `feat/issue-3-add-task-due-date`:** la funcionalidad ya está implementada y comprometida respecto a `main` (34 ficheros, 3189 inserciones). No hay cambios de código sin confirmar que pertenezcan al cambio.

Verificación contra los criterios:

| Criterio | Evidencia en la rama |
|---|---|
| 1, 3 | `models/task.go`: `DueDate *time.Time` (nullable). `Test_Task_DueDate_Optional`, `TestTasksCreate_Valid` comprueba `Nil(task.DueDate)`. |
| 2 | Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, tipo `date`, sin hora. `Test_Task_DueDate_DayOnly_Persists` comprueba hora 0. |
| 4 | `actions/tasks.go` `parseDueDate` con `time.Parse("2006-01-02")`. `templates/tasks/index.plush.html` y `edit.plush.html` usan `type="date"`. |
| 5 | `templates/tasks/index.plush.html`: `t.DueDate.Format("02/01/2006")`. `Test_TasksIndex_ShowsDueDate`. |
| 6 | `TasksCreate` y `TasksUpdate` añaden `due_date` a `verrs`; `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`. |
| 7 | `Test_TasksToggle_KeepsDueDate`. |
| 8 | `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`. |

Conclusión: no hay un criterio sin cubrir en código ni en tests. Lo que queda es validar, no implementar.

**Propuesta:** no añadir código de aplicación. Ejecutar la validación, registrar el resultado y preparar el PR con el cambio ya existente más este plan.

## Alcance

**Incluido:**

- Verificación del código y tests ya presentes en la rama.
- Ejecución de las validaciones del proyecto.
- Preparación del PR desde esta rama.

**Excluido:**

- Cambios en `adws/`, `config/adw.yml` (modificado localmente, ajeno al cambio) o en la lógica de disparo del agente.
- Retirar los planes antiguos de `plans/` (≈119 ficheros sin trackear, más los ya commiteados en la rama). Es una decisión del usuario.
- Campos de hora, zonas horarias, recordatorios, filtros por fecha o ordenación por fecha.

## Diseño por capas

Capas afectadas por la funcionalidad (ya implementadas):

- **Persistencia:** migración `date` nullable.
- **Modelo:** `models.Task.DueDate *time.Time`, `form:"due_date"`.
- **Acción:** `TasksCreate`, `TasksUpdate` (parseo y validación), `TasksEdit`/`TasksIndex` (valor de vuelta).
- **Renderizado:** `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`.
- **CSS:** `public/assets/app.css` (`.task-due`, `.new-task-row input[type="date"]`).

No hay capa JavaScript en el cambio.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y retira la columna `due_date` (`date`, nullable). |
| `models/task.go` | Campo `DueDate`. |
| `models/task_test.go` | Tests de modelo: opcional y persistencia solo día. |
| `actions/tasks.go` | `parseDueDate`, validación en create/update, `dueDateInput` al re-renderizar. |
| `actions/tasks_test.go` | Tests de acción: crear, editar, limpiar, inválida, toggle, render. |
| `actions/actions_test.go` | Helper `authed` con token CSRF. |
| `templates/tasks/index.plush.html` | Campo de fecha en el alta y etiqueta en la lista. |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición con ayuda y error. |
| `public/assets/app.css` | Estilos de la etiqueta de fecha y del campo en la fila de alta. |

Fuera del cambio (no tocar): `adws/*`, `config/adw.yml`, `migrations/schema.sql` (sin trackear, ajeno).

## Implementación paso a paso

1. Confirmar que la rama está en `feat/issue-3-add-task-due-date` y que el árbol no tiene cambios de código propios sin confirmar.
2. No modificar código. Si la validación de los pasos siguientes falla, detenerse y reportar.
3. Ejecutar las validaciones de la sección siguiente.
4. Para el PR: hacer stage solo de este plan y de los ficheros ya comprometidos en la rama. No añadir `config/adw.yml`, `.claude/skills/ship/`, `.txt`, `.x`, `migrations/schema.sql` ni el resto de planes sin trackear.
5. Hacer push a `origin` y abrir el PR contra `main`.

## Tests y validación

Condiciones: PostgreSQL escuchando en `127.0.0.1:5432` (comprobado) con la base `todo_test` (usuario `postgres`). Los tests de aplicación requieren `GO_ENV=test` para no escribir en `development` (ver memoria del proyecto).

Resultados observados al preparar este plan:

| Comando | Resultado |
|---|---|
| `go vet ./...` | OK (sin salida). |
| `go build ./...` | OK (sin salida). |
| `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/` | OK (`todo/actions` 0.459s, `todo/models` 0.322s). |
| `GO_ENV=test go test -p 1 ./...` | FALLA en `todo/adws` (5 tests de `TestTrigger*`), no relacionados. |
| `gofmt -l .` | Lista casi todos los ficheros por CRLF (ver riesgos). |
| `gofmt -l` sobre el contenido normalizado a LF de los ficheros tocados | OK (sin salida). |

Lectura de los fallos de `adws`: los tests usan un repositorio fijo `example/todo` y llaman a `gh`, que devuelve "Could not resolve to a Repository". `adws/` no tiene diferencias respecto a `main`. No se ha comprobado si fallan también en `main`; se deduce que son previos al cambio.

Criterio de paso para el PR: `go vet`, `go build`, y los tests de `actions` y `models` en verde. Los fallos de `adws` se mencionan en el PR como preexistentes y no bloquean, salvo que el usuario decida otra cosa.

## Riesgos y recuperación

- **CRLF en el árbol de trabajo:** `core.autocrlf=true` convierte el checkout a CRLF, mientras que el repositorio guarda LF. `gofmt -l .` reporta todos los ficheros aunque el código esté bien formateado. No es un fallo del cambio. Para validar el formato usar la comprobación sobre el contenido normalizado.
- **Planes en el PR:** la rama ya incluye ≈30 ficheros `plans/issue-3-*.md` comprometidos desde `main`. Van en el PR tal cual. Si no se quieren, hay que decidirlo antes de abrir el PR (reescribir la rama queda fuera de este plan, porque implica un push forzado).
- **Cambios ajenos sin confirmar:** `config/adw.yml` está modificado localmente (repo y prompt del trigger). No es parte de la funcionalidad y no debe entrar en el PR.
- **Tests de `adws`:** si el usuario exige suite verde completa, hace falta arreglar `adws` por separado, fuera de esta issue.
- **Recuperación:** si el PR se abre y hay que retirarlo, el cambio se puede revertir con un commit nuevo; no hay migraciones destructivas (el `down` solo elimina la columna `due_date`, que solo existe en esta rama).

## Siguiente paso

`/implement plans/issue-3-fecha-limite-solo-dia-plan-ship-verificacion-2026-10-23.md`
