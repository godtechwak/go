package main

import (
    "database/sql"
    "fmt"
    "os"
    "strconv"
    "time"

    "context"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/aws/aws-sdk-go/service/rds"
)

func main() {
    num, _ := strconv.ParseInt(os.Args[1], 10, 64)

    for {
        currentTime := time.Now()
        hour, min, sec := currentTime.Clock()
        millisec := currentTime.Nanosecond() / 1000000
        timeString := fmt.Sprintf("%02d:%02d:%02d:%03d", hour, min, sec, millisec)

        // Create DB object
        db, err := sql.Open("mysql", "{사용자명}:{비밀번호}@tcp({서버주소}:3306)/mysql")
        if err != nil {
            fmt.Printf("%s create db object error(%s)\n", timeString, err)
            continue
        }

        // Ping check
        ctx := context.Background()
        err = db.PingContext(ctx)
        if err != nil {
            fmt.Printf("%s Can not connect(%s)\n", timeString, err)
            time_duration(currentTime, num)
            continue
        }

        // Execute Query
        rows, err := db.Query("SELECT @@hostname as hostname")
        if err != nil {
            fmt.Printf("%s Can not query(%s)\n", timeString, err)
            time_duration(currentTime, num)
            continue
        }

        // Execute insert
        /* CREATE TABLE test (id auto_increment primary key, created_at datetime)
        insert, err := db.Query("INSERT INTO silver.test SELECT null, now(3)")
        if err != nil {
            fmt.Printf("%s Can not insert(%s)\n", timeString, err)
            time_duration(currentTime, num)
            continue
        }
        insert.Close()
        */

        // Copy rows
        for rows.Next() {
            var hostname string
            err := rows.Scan(&hostname)
            if err != nil {
                fmt.Printf("%s copy rows error(%s)\n", timeString, err)
                continue
            }
            fmt.Printf("%s OK from %s\n", timeString, hostname)
        }
        time_duration(currentTime, num)
        db.Close()
    }
}

func time_duration(currentTime time.Time, num int64) {
    endMillis := time.Now()
    diffMillis := num + currentTime.Sub(endMillis).Milliseconds()
    if (diffMillis * (-1)) < num {
        time.Sleep(time.Duration(diffMillis) * time.Millisecond)
    } else {
        time.Sleep(time.Duration(num) * time.Millisecond)
    }
}
