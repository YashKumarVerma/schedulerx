# Schedulerx 
> The objective of this project is to implement a job scheduler that will run various tasks (commands) based on their cron schedules. The scheduler must be scalable and function correctly in a multi-pod environment, ensuring that each scheduled command runs exactly once at the scheduled time, even if multiple servers are executing the scheduler service.


### Interfaces and Components:

1. **Command Interface**:

   - `Command` interface represents a unit of work that the `schedulerx` package will schedule.
   - Each `Command` should have a unique identifier.
   - The `Command` interface will allow users of the scheduler to register different types of commands that can be scheduled for execution.

   **Example**:

   ```pseudo
   interface Command {
       string ID();           // Unique ID of the command
       error Execute(params []string);   // Method to execute the command
   }
   ```

2. **ScheduleFetcher Interface**:

   - Define a `ScheduleFetcher` interface responsible for retrieving the latest schedule for a registered `Command`.
   - The schedule will follow cron format.
   - The `ScheduleFetcher` should also provide parameters that will be passed to the Command while executing it.

   **Example**:

   ```pseudo
   interface ScheduleFetcher {
       (string, []string, error) FetchSchedule(string commandID);  // Fetches the cron schedule and params for the given command
   }
   ```


### Functionality:

- The job of the scheduler is to:
  1. Periodically fetch the latest schedules from the `ScheduleFetcher` for the registered commands.
  2. Execute the commands with the correct parameters at the specified cron times.
  3. Ensure commands run exactly once at the scheduled time, even in a multi-pod environment (i.e., when the scheduler service is running on multiple servers).
  4. The scheduler should be able to return a list of Jobs (A job is a single execution of a Command). Every job should have a unique id along with a status (running, scheduled, failed). For every Command registered in the scheduler, it should always have the next 2 instances of 'Job' scheduled. i.e., for every command, the scheduler should be able to list down all the executions (i.e., Jobs) that have completed and the next 2 scheduled Jobs.
  5. You can use any library to parse the crontab.
  6. For now, implement a basic version of ScheduleFetcher and inject it into the scheduler.

### Expected Deliverables:

- Define all necessary models, structs, and services to make the scheduler package work, including but not limited to:
  - Command registration
  - Schedule fetching and execution logic
  - Handling multi-pod environments for safe execution
