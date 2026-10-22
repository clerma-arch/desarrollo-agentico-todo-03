# Chore: añadir un test para la validación de título vacío en una tarea

## Encargo y resultado esperado

Añadir un test que demuestre la validación de título vacío en una tarea. Resultado: un nuevo test en `models/task_test.go` que pasa con el código actual y falla si se elimina la regla de título obligatorio al actualizar. No se cambia código de producción.

## Estado actual y restricciones

- `models/task.go` (`Validate`, líneas ~32-36): recorta con `strings.TrimSpace`; si queda vacío añade el error `"El título es obligatorio."` al campo `title`.
- Ya existen tests de título vacío:
  - Modelo, creación: `Test_Task_Title_Required` y `Test_Task_Title_OnlySpaces_Invalid` (`models/task_test.go`).
  - Acciones: `Test_TasksCreate_EmptyTitle_Invalid`, `Test_TasksCreate_OnlySpaces_Invalid`, `Test_TasksUpdate_EmptyTitle`, `Test_TasksUpdate_OnlySpaces` (`actions/tasks_test.go`).
- **Hueco detectado**: no hay un test a nivel de modelo que valide el título vacío al **actualizar** una tarea ya persistida (`ValidateAndUpdate`) ni que compruebe que el título guardado no se sobrescribe. Ese es el test mínimo no redundante a añadir.
- Restricciones: usar la suite `ModelSuite` existente (`ms.DB`), mensajes en español, sin dependencias nuevas. Los tests necesitan base de datos (`go test -p 1 ./...`).

## Alcance

Incluido:
- Un único test nuevo en `models/task_test.go`.

Excluido:
- Cambios en `models/task.go`, acciones, plantillas, fixtures o configuración.
- Duplicar los tests de creación o de acciones ya existentes.
- Nuevas herramientas o dependencias.

## Archivos relevantes

- `models/task.go`: regla de validación (solo lectura).
- `models/task_test.go`: aquí se añade el test; sigue el patrón de `Test_Task_Toggle_Persists` para crear y recargar.
- `actions/tasks_test.go`: referencia de cobertura existente (solo lectura).

## Implementación paso a paso

1. Antes de cambiar nada, ejecutar la línea base: `go test -p 1 ./models/...` (debe pasar).
2. En `models/task_test.go`, tras `Test_Task_Title_OnlySpaces_Invalid`, añadir:

```go
func (ms *ModelSuite) Test_Task_Title_Empty_OnUpdate_Invalid() {
	task := &Task{Title: "Llamar al banco"}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	task.Title = ""
	verrs, err = ms.DB.ValidateAndUpdate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.Equal("Llamar al banco", reloaded.Title)
}
```

3. No importar nada nuevo (el archivo ya importa `strings`; el test no lo necesita).

## Tests y validación

Orden:
1. `gofmt -l .` (sin salida)
2. `go vet ./...`
3. `go test -p 1 ./models/... -run Test_Task_Title` y después `go test -p 1 ./...`
4. `go build ./...`

Comprobación de que el test es significativo (opcional, temporal): comentar la rama `title == ""` en `models/task.go`, verificar que el test falla y restaurar el archivo (`git diff models/task.go` debe quedar vacío).

Comportamiento a preservar: todos los tests previos siguen pasando sin modificación.

## Riesgos y recuperación

- Si `ValidateAndUpdate` no devuelve el error esperado, indicaría un bug real: detenerse y reportarlo, sin arreglar en este chore.
- Requiere la base de datos de test disponible; si no lo está, los tests fallarán por conexión, no por el cambio.
- Recuperación: `git checkout -- models/task_test.go`.

## Ajuste posterior (decisión del usuario: sin base de datos)

El test propuesto con `ValidateAndUpdate` requiere BD y se sustituye por `Test_Task_Validate_EmptyTitle` en `models/task_test.go`: test estándar (`testing`) que llama a `task.Validate(nil)` con `""` y solo espacios/tabulador y comprueba el error `"El título es obligatorio."` en `title`. No usa `ModelSuite` ni Postgres. Validación: `go test ./models/ -run Test_Task_Validate_EmptyTitle -v`.
