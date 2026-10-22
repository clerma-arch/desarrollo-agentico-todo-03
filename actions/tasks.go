package actions

import (
	"net/http"
	"strings"
	"time"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
)

const dueDateInvalidMessage = "La fecha no es válida."

// parseDueDate convierte la fecha del formulario (YYYY-MM-DD) en un día sin hora.
// Una cadena vacía significa que la tarea no tiene fecha.
func parseDueDate(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}

	due, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &due, nil
}

// TasksIndex muestra la lista de tareas y el formulario de creación.
func TasksIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	tasks := models.Tasks{}
	if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
		return err
	}

	c.Set("tasks", tasks)
	c.Set("task", &models.Task{})
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/index.plush.html"))
}

// TasksCreate crea una tarea nueva a partir del formulario.
func TasksCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	input := struct {
		Title   string `form:"title" json:"title"`
		DueDate string `form:"due_date" json:"due_date"`
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}

	task := &models.Task{Title: strings.TrimSpace(input.Title)}
	dueDate, dueErr := parseDueDate(input.DueDate)
	task.DueDate = dueDate

	verrs, err := task.Validate(tx)
	if err != nil {
		return err
	}
	if dueErr != nil {
		verrs.Add("due_date", dueDateInvalidMessage)
	}

	if verrs.HasAny() {
		tasks := models.Tasks{}
		if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
			return err
		}

		c.Set("tasks", tasks)
		c.Set("task", task)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/index.plush.html"))
	}

	if err := tx.Create(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea creada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksEdit muestra el formulario de edición de una tarea existente.
func TasksEdit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("task", task)
	c.Set("errors", validate.NewErrors())
	return c.Render(http.StatusOK, r.HTML("tasks/edit.plush.html"))
}

// TasksUpdate actualiza el título y la fecha límite, conservando los demás datos de la tarea.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	input := struct {
		Title   string `form:"title" json:"title"`
		DueDate string `form:"due_date" json:"due_date"`
	}{}
	if err := c.Bind(&input); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(input.Title)
	dueDate, dueErr := parseDueDate(input.DueDate)
	task.DueDate = dueDate

	verrs, err := task.Validate(tx)
	if err != nil {
		return err
	}
	if dueErr != nil {
		verrs.Add("due_date", dueDateInvalidMessage)
	}
	if verrs.HasAny() {
		c.Set("task", task)
		c.Set("errors", verrs)
		return c.Render(http.StatusUnprocessableEntity, r.HTML("tasks/edit.plush.html"))
	}

	if err := tx.Update(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea actualizada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksToggle alterna una tarea entre completada y pendiente.
func TasksToggle(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	task.Completed = !task.Completed
	if err := tx.Update(task); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksDestroy elimina una tarea.
func TasksDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea eliminada.")
	return c.Redirect(http.StatusSeeOther, "/")
}
