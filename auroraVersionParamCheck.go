package main

import (
    "fmt"
    "log"
    "time"
    "os"
    "strconv"
    "text/tabwriter"
    "io/ioutil"
    "strings"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/rds"
  _ "github.com/go-sql-driver/mysql"

)

func readInput(stage string) string {
    for {
        if stage == "region" {
            fmt.Printf("Region(kr/jp/ca/uk): \n")
            var region string
            _, err_region := fmt.Scan(&region)

            if err_region != nil {
                fmt.Printf("Input Data Error: %v\n", err_region)
                continue
            }

            if region == "kr" || region == "jp" || region == "ca" || region == "uk" {
                return region
            } else {
                fmt.Printf("Input correct Region(%s)\n", region)
                continue
            }
        } else if stage == "worktype" {
            fmt.Printf("WorkType(cluster/instance): \n")
            var worktype string
            _, err_worktype := fmt.Scan(&worktype)

            if err_worktype != nil {
                fmt.Printf("Input Data Error: %v\n", err_worktype)
                continue
            }

            if worktype == "cluster" || worktype == "instance" {
                return worktype
            } else {
                fmt.Printf("Input correct WorkType(%s)\n", worktype)
                continue
            }
        }
    }
}

func auroraVersionParamCluster(svc *rds.RDS, lines []string, w *tabwriter.Writer, num int64) {
    init_map := make(map[string]string) // 수행시간을 담기 위한 맵
    var count int //반복횟수
    count = 1
    var duration time.Duration //수행시간


    for {
        fmt.Fprintf(w, "──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t\n")
        fmt.Fprintf(w, " Time│\t Duration|\t Cluster│\t Version│\t Status│\t Param Status│\t\n")
        fmt.Fprintf(w, "──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t\n")
        for _, rcluster_name := range lines {
            currentTime := time.Now()
            hour, min, sec := currentTime.Clock()
            millisec := currentTime.Nanosecond() / 1000000

            timeFormat := "15:04:05.000"
            timeString := fmt.Sprintf("%02d:%02d:%02d.%03d", hour, min, sec, millisec)

            if count == 1 {
              init_map[rcluster_name] = timeString //최초 수행된 시간을 맵에 담아놓는다.
            } else {
              first_timeString, exists := init_map[rcluster_name]

              if exists {
                first_time, err_firsttime := time.Parse(timeFormat, first_timeString)
                after_time, err_aftertime := time.Parse(timeFormat, timeString)

                if err_firsttime != nil || err_aftertime != nil {
                  fmt.Printf("Parsing Error: %s %s", err_firsttime, err_aftertime)
                  return
                }

                duration = after_time.Sub(first_time) //마지막에 수행된 시간에서 최초 수행된 시간의 차를 계산한다.
              }
            }

            input := &rds.DescribeDBClustersInput{
                    DBClusterIdentifier: aws.String(rcluster_name),
                }

            result, err := svc.DescribeDBClusters(input)

            if err != nil {
                fmt.Println("DBCluster Error: ", err)
                return
            }

            // 클러스터 기본 정보(클러스터명, 엔진 버전, 클러스터 상태)
            cluster_info := result.DBClusters[0]
            // DB 클러스터 파라미터 그룹 상태
            cluster_param := result.DBClusters[0].DBClusterMembers[0]

            // DB 클러스터 및 클러스터 파라미터 정보 출력
            fmt.Fprintf(w, "%s│\t %s|\t %s│\t %s│\t %s│\t %s│\t\n", timeString, duration, *cluster_info.DBClusterIdentifier, *cluster_info.EngineVersion, *cluster_info.Status, *cluster_param.DBClusterParameterGroupStatus)

        }

        fmt.Fprintf(w, "──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t\n")
        fmt.Print("\033[2J\033[H")
        w.Flush()
        time.Sleep(time.Duration(num) * time.Millisecond)
        count += 1
    }
}

func auroraVersionParamInstance(svc *rds.RDS, lines []string, w *tabwriter.Writer, num int64) {
  init_map := make(map[string]string) // 수행시간을 담기 위한 맵
    var count int //반복횟수
    count = 1
    var duration time.Duration //수행시간

    for {
        fmt.Fprintf(w, "──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t\n")
        fmt.Fprintf(w, " Time│\t Duration│\t Instance│\t Version│\t Status│\t Param Status│\t\n")
        fmt.Fprintf(w, "──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t\n")

        for _, rinstance_name := range lines {
            currentTime := time.Now()
            hour, min, sec := currentTime.Clock()
            millisec := currentTime.Nanosecond() / 1000000

            timeFormat := "15:04:05.000"
            timeString := fmt.Sprintf("%02d:%02d:%02d.%03d", hour, min, sec, millisec)

            if count == 1 {
              init_map[rinstance_name] = timeString //최초 수행된 시간을 맵에 담아놓는다.
          } else {
              first_timeString, exists := init_map[rinstance_name]

              if exists {
                first_time, err_firsttime := time.Parse(timeFormat, first_timeString)
                after_time, err_aftertime := time.Parse(timeFormat, timeString)

                if err_firsttime != nil || err_aftertime != nil {
                  fmt.Printf("Parsing Error: %s %s", err_firsttime, err_aftertime)
                  return
                }

                duration = after_time.Sub(first_time) //마지막에 수행된 시간에서 최초 수행된 시간의 차를 계산한다.
              }
            }

            input := &rds.DescribeDBInstancesInput{
                DBInstanceIdentifier: aws.String(rinstance_name),
            }

            result, err := svc.DescribeDBInstances(input)

            if err != nil {
                fmt.Println("Error", err)
                return
            }

            // 인스턴스 기본 정보(인스턴스명, 엔진 버전, 인스턴스 상태)
            instance_info := result.DBInstances[0]
            // DB 인스턴스 파라미터 그룹 상태
            instance_param := result.DBInstances[0].DBParameterGroups[0]

            // DB 인스턴스 및 인스턴스 파라미터 정보 출력
            fmt.Fprintf(w, "%s│\t %s│\t %s│\t %s│\t %s│\t %s│\t\n", timeString, duration, *instance_info.DBInstanceIdentifier, *instance_info.EngineVersion, *instance_info.DBInstanceStatus, *instance_param.ParameterApplyStatus)
        }

        fmt.Fprintf(w, "──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t ──────────────────────────────────\t\n")
        fmt.Print("\033[2J\033[H")
        w.Flush()
        time.Sleep(time.Duration(num) * time.Millisecond)
        count += 1
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Please provide a number as argument.")
        return
    }

    num, err := strconv.ParseInt(os.Args[1], 10, 64)
    if num < 100 {
        fmt.Println("Input more than 100 milliseconds")
        return
    }

    if err != nil {
        fmt.Printf("Error parsing number: %v", err)
        return
    }

    var real_region string

    w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight)

    fmt.Printf("================================\n")
    fmt.Printf("Aurora Version & Parameter Check\n")
    fmt.Printf("================================\n")

    region := readInput("region")
    worktype := readInput("worktype")

    regionMap := map[string]string {
        "kr": "ap-northeast-2",
        "jp": "ap-northeast-1",
        "ca": "ca-central-1",
        "uk": "eu-west-2",
    }

    real_region, ok := regionMap[region]

    if !ok {
        fmt.Println("Invalid region provided")
        return
    }

    sess, err := session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
        Config: aws.Config{
            Region: aws.String(real_region),
        },
    })

    if worktype == "cluster" {
        filePath := "./db_cluster_list_" + region + ".txt"
        content, err := ioutil.ReadFile(filePath)

        if err != nil {
            log.Fatal(err)
        }
        cluster_name := string(content)
        lines := strings.Split(cluster_name, "\n")

        svc := rds.New(sess)

        auroraVersionParamCluster(svc, lines, w, num)

    } else if worktype == "instance" {
        filePath := "./db_instance_list_" + region + ".txt"
        content, err := ioutil.ReadFile(filePath)

        if err != nil {
            log.Fatal(err)
        }

        instance_name := string(content)
        lines := strings.Split(instance_name, "\n")

        svc := rds.New(sess)

        auroraVersionParamInstance(svc, lines, w, num)
    }
}
