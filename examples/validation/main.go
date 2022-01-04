// Example program for validation, showcasing the following:
// Validation of parameters as a dependent, reusable task.
// Execute `go run . -v migrate -db only-up -direction up` as a valid example.
// Execute `go run . -v migrate -db only-up -direction down` as an invalid example.
// Execute `go run . -h` to see all details.

package main

import (
	"github.com/goyek/goyek"
)

func main() {
	flow := &goyek.Flow{}

	db := flow.RegisterStringParam(goyek.StringParam{
		Name:  "db",
		Usage: "The database to migrate ('only-up' / 'only-down')",
	})
	direction := flow.RegisterStringParam(goyek.StringParam{
		Name:  "direction",
		Usage: "The direction to migrate ('up' / 'down')",
	})

	validation := flow.Register(taskValidation(db, direction))
	flow.Register(taskMigrate(validation, db, direction))

	flow.Main()
}

// taskValidation validates params and prints error messages for each error
// it encounters. If it encounters at least one error, the validation function
// will abort the TaskFlow and the task will be stopped.
func taskValidation(db, direction goyek.RegisteredStringParam) goyek.Task {
	return goyek.Task{
		Name:   "validation",
		Params: goyek.Params{db, direction},
		Action: func(tf *goyek.TF) {
			dbStr := db.Get(tf)
			dirStr := direction.Get(tf)
			if dbStr != "only-up" && dbStr != "only-down" {
				tf.Errorf("Database %q is invalid", dbStr)
			}
			if dirStr != "up" && dirStr != "down" {
				tf.Errorf("Direction %q is invalid", dirStr)
			}
			if (dbStr == "only-up" && dirStr == "down") || (dbStr == "only-down" && dirStr == "up") {
				tf.Errorf("Direction %q is invalid for database %q", dirStr, dbStr)
			}
		},
	}
}

// taskMigrate prints a message with the provided params.
func taskMigrate(validation goyek.RegisteredTask, db, direction goyek.RegisteredStringParam) goyek.Task {
	return goyek.Task{
		Name:   "migrate",
		Usage:  "Migrate a database",
		Params: goyek.Params{db, direction},
		Deps:   goyek.Deps{validation},
		Action: func(tf *goyek.TF) {
			tf.Logf("Database %q and direction %q are valid values", db.Get(tf), direction.Get(tf))
		},
	}
}
