# Greener
Greener is a lean and mean test result explorer.

Among other use cases, it lets you:

- Get to test results fast (query specific sessions, tests, statuses, labels etc.)
- Group test results and check aggregated statuses (e.g. `group_by(#"os", #"version")` labels)

Features:

- Easy to use
- No changes to test code needed
- Simple SQL-like query language (with grouping support)
- Attach labels and/or baggage (arbitrary JSON) to test sessions
- Self-contained executable (only requires SQLite/PostgreSQL/MySQL database)
- Small (~27mb executable / compressed Docker image)

Demo:
![Demo](assets/demo.gif)
