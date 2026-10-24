# Feature: fecha límite opcional (solo día) en las tareas — validación y PR (issue #3)

## Encargo y criterios de aceptación

Encargo (issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"):

- Cada tarea puede tener una fecha límite opcional.
- La fecha indica solo el día, sin hora.

Criterios de aceptación (requisitos explícitos marcados con [E], inferencias con [I]):

1. [E] La fecha límite es opcional: crear una tarea sin fecha la deja sin fecha.
2. [E] Solo se guarda el día (sin hora).
3. [I] Se puede indicar la fecha al crear y al editar una tarea.
4. [I] La lista de tareas muestra la fecha límite cuando existe.
5. [I] Una fecha mal formada devuelve 422 y no guarda nada.
6. [I] Si falla la validación (título vacío o fecha inválida), el formulario conserva la fecha escrita.
7. [I] Editar solo el título conserva la fecha; completar una tarea también la conserva.
8. [I] Borrar la fecha en la edición la deja sin valor.

## Comportamiento actual y propuesta

Estado actual en la rama `feat/issue-3-add-task-due-date` (HEAD `b5c207c`). Ya existe implementación:

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `down`.
- Modelo `models/task.go`: campo `DueDate *time.Time` con etiquetas `json`, `db` y `form`.
- Acciones `actions/tasks.go`: `parseDueDate` (acepta `YYYY-MM-DD` o vacío), y `TasksCreate` / `TasksUpdate` que añaden el error `due_date` y re-renderizan con 422. `TasksEdit` e `TasksIndex` reinician `dueDateInput`.
- Plantillas `templates/tasks/index.plush.html` y `templates/tasks/edit.plush.html`: input `type="date"` `name="due_date"`, texto de ayuda, error con `role="alert"` y mostrado de la fecha en la lista (`02/01/2006`).
- CSS en `public/assets/app.css` (11 líneas añadidas).
- Tests en `actions/tasks_test.go` y `models/task_test.go` que cubren alta, edición, fecha inválida, conservación de la fecha al fallar validación, al completar y al editar solo el título.

Propuesta: no reescribir la funcionalidad. El trabajo de este plan es verificar que el estado actual cumple los criterios, ejecutar la validación completa y preparar el PR con solo los archivos del cambio y de este plan.

## Alcance

Incluido:

- Ejecutar validación (gofmt, vet, tests, build) sobre la rama tal como está.
- Revisar que los tests cubren los criterios 1–8. Si falta un test para un criterio, añadirlo siguiendo el patrón de `actions/tasks_test.go`.
- Commit(s) mínimos solo si la validación exige un cambio, y push a `origin`.
- Abrir PR contra `main`.

Excluido:

- Cambiar el comportamiento de la fecha (formato, zona horaria, hora, recordatorios, filtros).
- Tocar `config/adw.yml` (cambio previo sin confirmar).
- Incluir `migrations/schema.sql`, `.txt`, `.x`, `.claude/skills/ship/` o carpetas no relacionadas.
- Hacer merge o `--force`.

## Diseño por capas

- Persistencia y migraciones: ya existe `20261022120000_add_due_date_to_tasks`. No añadir otra migración salvo que la validación demuestre que falta.
- Modelo: `models/task.go`, sin cambios previstos.
- Acción y ruta: `actions/tasks.go`, `TasksCreate` y `TasksUpdate`. Sin cambios previstos salvo un fallo detectado en validación.
- Renderizado: `templates/tasks/index.plush.html` y `templates/tasks/edit.plush.html`. Sin cambios previstos.
- CSS: `public/assets/app.css`, solo la clase `.task-due` y el hint. Sin cambios previstos.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y quita la columna `due_date` (date, nullable) |
| `models/task.go` | Campo `DueDate *time.Time` |
| `actions/tasks.go` | `parseDueDate`, alta, edición y validación con 422 |
| `actions/tasks_test.go` | Tests de comportamiento de la fecha en acciones |
| `models/task_test.go` | Tests de modelo |
| `templates/tasks/index.plush.html` | Campo de fecha en alta, lista con fecha |
| `templates/tasks/edit.plush.html` | Campo de fecha en edición |
| `public/assets/app.css` | Estilos de la fecha |
| `plans/issue-3-fecha-limite-dia-validacion-y-pr-2026-10-24.md` | Este plan |

## Implementación paso a paso

1. Confirmar que el árbol de trabajo solo tiene `config/adw.yml` y los `??` no relacionados: `git status --short --branch`.
2. Ejecutar la validación de la sección siguiente. Si todo pasa, no tocar código.
3. Revisar que cada criterio [E]/[I] tiene un test que lo demuestra (tabla en "Tests y validación"). Si falta, añadir el test en `actions/tasks_test.go` con el patrón `as.authed(...)` existente.
4. Si un test falla, detenerse y reportar la salida. No ampliar el alcance.
5. Stage explícito solo de los archivos del cambio y de este plan. Nunca `git add .` ni `git add -A`. Lista prevista: los archivos de la sección "Archivos relevantes" que estén modificados o nuevos en el diff contra `main`, más `plans/issue-3-fecha-limite-dia-validacion-y-pr-2026-10-24.md`.
6. Si hay cambios de código, commit Conventional Commits con `Co-Authored-By: Claude Haiku 5.5 <noreply@anthropic.com>`.
7. `git push -u origin feat/issue-3-add-task-due-date`.
8. `gh pr create --base main` con título Conventional Commits (`feat(tasks): añadir fecha límite opcional (solo día) a las tareas`) y descripción con resumen, ruta del plan, tests y resultado de validación, terminando con `🤖 Generated with [Claude Code](https://claude.com/claude-code)`.

## Tests y validación

Comandos (desde la raíz del repo, PowerShell o Git Bash):

- `gofmt -l .` → debe salir vacío.
- `go vet ./...` → sin errores.
- `GO_ENV=test go test -p 1 ./...` → todo en verde. `GO_ENV=test` es obligatorio: sin él, los tests de `actions/` escriben en la base de desarrollo.
- `go build ./...` → sin errores.

Prerrequisitos: PostgreSQL accesible con la configuración de `config/database.yml` para el entorno `test`, y las migraciones de test aplicadas (incluida `20261022120000`).

Cobertura por criterio (tests ya existentes en `actions/tasks_test.go`):

| Criterio | Test |
|---|---|
| 1 Opcional | `TasksCreate` sin `due_date` (cubierto por el test de alta base, `as.Nil(task.DueDate)`) |
| 2, 3 Alta con día | `Test_TasksCreate_WithDueDate` |
| 5 Fecha inválida 422 | `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate` |
| 6 Conserva fecha al fallar | `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate` |
| 4 Muestra en lista | `Test_TasksIndex_ShowsDueDate` |
| 3 Edición y prefill | `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate` |
| 7 Conserva al completar / editar título | `Test_TasksToggle_KeepsDueDate`, `Test_TasksUpdate_KeepsDueDateWhenOnlyTitleChanges` |
| 8 Borrar fecha | `Test_TasksUpdate_SetsAndClearsDueDate` |

Si `grep` no encuentra alguno de estos nombres en el momento de ejecutar, el criterio correspondiente no está cubierto y hay que añadir el test (paso 3).

## Riesgos y recuperación

- Estado no ejecutado: en la fase de planificación no se han corrido tests. El resultado real solo se conoce al ejecutar la validación; reportar la salida tal cual.
- Edición vía JSON sin `due_date`: `TasksUpdate` asigna `task.DueDate = dueDate` aunque el campo venga vacío, así que una petición sin la clave borra la fecha. El formulario HTML siempre envía el campo, por lo que no afecta a la UI; queda como comportamiento a documentar, no a cambiar aquí.
- Zona horaria: la fecha se parsea en UTC (`time.Parse` sin zona). Los tests comparan `Format("2006-01-02")`, por lo que son estables; un cambio de zona no está en el alcance.
- Volumen de `plans/`: la rama tiene decenas de planes de esta issue sin integrar. El PR debe incluir solo el plan de este ciclo, no el resto.
- Archivos sin relación: `config/adw.yml`, `migrations/schema.sql` (sin seguimiento), `.txt`, `.x`, `.claude/skills/ship/`. Quedan fuera del PR.
- Recuperación: si la validación falla, no abrir PR. Revertir solo el commit de test o fix introducido en este ciclo (`git revert <sha>`, nunca `reset --hard` ni `--force`).
