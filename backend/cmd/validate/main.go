package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"indoor-navigation/backend/internal/graph"
	"indoor-navigation/backend/internal/repository"
)

func main() {
	var (
		dsn      string
		building string
		asJSON   bool
		strict   bool
	)

	flag.StringVar(&dsn, "dsn", os.Getenv("MYSQL_DSN"), "MySQL DSN，默认读取 MYSQL_DSN")
	flag.StringVar(&building, "building", "C5", "楼栋编码")
	flag.BoolVar(&asJSON, "json", false, "以 JSON 输出校验报告")
	flag.BoolVar(&strict, "strict", true, "警告视为错误")
	flag.Parse()

	if strings.TrimSpace(dsn) == "" {
		fmt.Fprintln(os.Stderr, "缺少 MYSQL_DSN，或使用 -dsn 指定连接串")
		os.Exit(2)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接 MySQL 失败: %v\n", err)
		os.Exit(2)
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取数据库连接失败: %v\n", err)
		os.Exit(2)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "MySQL Ping 失败: %v\n", err)
		os.Exit(2)
	}

	raw, err := repository.NewGraphRepository(db).Load(ctx, building)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取图数据失败: %v\n", err)
		os.Exit(2)
	}

	report := graph.Validate(raw, strict)
	if asJSON {
		data, marshalErr := json.MarshalIndent(report, "", "  ")
		if marshalErr != nil {
			fmt.Fprintf(os.Stderr, "序列化校验报告失败: %v\n", marshalErr)
			os.Exit(2)
		}
		fmt.Println(string(data))
	} else {
		printReport(report)
	}

	if !report.Valid {
		os.Exit(1)
	}
}

func printReport(report graph.Report) {
	fmt.Printf("楼栋: %s\n", report.Building)
	fmt.Printf(
		"数据: floors=%d rooms=%d facilities=%d nodes=%d edges=%d classrooms=%d\n",
		report.Counts.Floors,
		report.Counts.Rooms,
		report.Counts.Facilities,
		report.Counts.Nodes,
		report.Counts.Edges,
		report.Counts.Classrooms,
	)
	if report.Valid {
		fmt.Println("结果: 通过")
	} else {
		fmt.Println("结果: 未通过")
	}

	if len(report.Errors) > 0 {
		fmt.Println("\n错误:")
		for _, issue := range report.Errors {
			printIssue(issue)
		}
	}
	if len(report.Warnings) > 0 {
		fmt.Println("\n警告:")
		for _, issue := range report.Warnings {
			printIssue(issue)
		}
	}
}

func printIssue(issue graph.Issue) {
	if issue.ID != "" {
		fmt.Printf("  [%s] %s(%s): %s\n", issue.Rule, issue.Table, issue.ID, issue.Message)
		return
	}
	fmt.Printf("  [%s] %s: %s\n", issue.Rule, issue.Table, issue.Message)
}
