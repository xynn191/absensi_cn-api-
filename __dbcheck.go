package main

import (
  "database/sql"
  "fmt"
  "log"

  "absensi-cn-api/internal/config"
  _ "github.com/go-sql-driver/mysql"
)

func main() {
  cfg, err := config.Load()
  if err != nil { log.Fatal(err) }
  dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.Params)
  db, err := sql.Open("mysql", dsn)
  if err != nil { log.Fatal(err) }
  defer db.Close()

  rows, err := db.Query("SHOW TABLES")
  if err != nil { log.Fatal(err) }
  defer rows.Close()

  for rows.Next() {
    var name string
    if err := rows.Scan(&name); err != nil { log.Fatal(err) }
    fmt.Println(name)
  }
  if err := rows.Err(); err != nil { log.Fatal(err) }
}
