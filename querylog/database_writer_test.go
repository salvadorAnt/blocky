package querylog

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo/v2"
)

var err error

var _ = Describe("DatabaseWriter", func() {

	Describe("Database query log to sqlite", func() {
		var (
			sqliteDB gorm.Dialector
			writer   *DatabaseWriter
		)

		BeforeEach(func() {
			sqliteDB = sqlite.Open("file::memory:")
		})

		When("New log entry was created", func() {
			BeforeEach(func() {
				writer, err = newDatabaseWriter(sqliteDB, 7, time.Millisecond)
				Expect(err).Should(Succeed())
			})

			It("should be persisted in the database", func() {
				// one entry with now as timestamp
				writer.Write(&LogEntry{
					Start:      time.Now(),
					DurationMs: 20,
				})

				// one entry before 2 days
				writer.Write(&LogEntry{
					Start:      time.Now().AddDate(0, 0, -2),
					DurationMs: 20,
				})

				// force write
				writer.doDBWrite()

				// 2 entries in the database
				Eventually(func() int64 {
					var res int64
					result := writer.db.Find(&logEntry{})

					result.Count(&res)

					return res
				}, "5s").Should(BeNumerically("==", 2))

				// do cleanup now
				writer.CleanUp()

				// now only 1 entry in the database
				Eventually(func() (res int64) {
					result := writer.db.Find(&logEntry{})

					result.Count(&res)

					return res
				}, "5s").Should(BeNumerically("==", 2))
			})
		})

		When("There are log entries with timestamp exceeding the retention period", func() {
			BeforeEach(func() {
				writer, err = newDatabaseWriter(sqliteDB, 1, time.Millisecond)
				Expect(err).Should(Succeed())
			})

			It("these old entries should be deleted", func() {
				// one entry with now as timestamp
				writer.Write(&LogEntry{
					Start:      time.Now(),
					DurationMs: 20,
				})

				// one entry before 2 days -> should be deleted
				writer.Write(&LogEntry{
					Start:      time.Now().AddDate(0, 0, -2),
					DurationMs: 20,
				})

				// force write
				writer.doDBWrite()

				// 2 entries in the database
				Eventually(func() int64 {
					var res int64
					result := writer.db.Find(&logEntry{})

					result.Count(&res)

					return res
				}, "5s").Should(BeNumerically("==", 2))

				// do cleanup now
				writer.CleanUp()

				// now only 1 entry in the database
				Eventually(func() (res int64) {
					result := writer.db.Find(&logEntry{})

					result.Count(&res)

					return res
				}, "5s").Should(BeNumerically("==", 1))
			})
		})
	})

	Describe("Database query log fails", func() {
		When("mysql connection parameters wrong", func() {
			It("should be log with fatal", func() {
				_, err := NewDatabaseWriter("mysql", "wrong param", 7, 1)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix("can't create database connection"))
			})
		})

		When("postgresql connection parameters wrong", func() {
			It("should be log with fatal", func() {
				_, err := NewDatabaseWriter("postgresql", "wrong param", 7, 1)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix("can't create database connection"))
			})
		})

		When("invalid database type is specified", func() {
			It("should be log with fatal", func() {
				_, err := NewDatabaseWriter("invalidsql", "", 7, 1)
				Expect(err).Should(HaveOccurred())
				Expect(err.Error()).Should(HavePrefix("incorrect database type provided"))
			})
		})
	})

})
