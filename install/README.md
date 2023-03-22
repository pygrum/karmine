# Installation

- Run `./install.sh`.

If something ever happens to your database instance(s), there is an initialisation script `init.sql` which will let you create a new SQLite3 database.  

```
sqlite3 karmine.db < init.sql
```