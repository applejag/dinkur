# Away detection

> This is a feature heavily inspired by
> [Grindstone](https://epiforge.com/grindstone).

The [Dinkur daemon](Glossary.md#dinkur-daemon) hooks into the operating system
and desktop environments to try and detect when the end-user has been away from
their computer.

This includes:

- if the end-user has been idle for a longer period (no mouse or keyboard input
  in a long while),

- locked their computer,

- if the computer has been hibernating,

- or if the computer has been rebooted.

If any of the above, then the Dinkur daemon tells the Dinkur clients to perform
an "away evaluation", to ask the user what to do with the time they have been
away. Evaluation actions are chosen by the end-user, and includes:

- Include the time spent away in the current task.
- Discard the time spent away.
- Use a different task name for the time spent away.
- Split the time spent away into multiple tasks.

As well as the option to continue or stop the task that was active when they
left.
