package opensearch

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewRepositoryFactory(t *testing.T) {
	Convey("TestNewRepositoryFactory", t, func() {
		Convey("成功创建 RepositoryFactory", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			So(factory, ShouldNotBeNil)
			So(factory.client, ShouldEqual, client)
		})

		Convey("使用 nil client 创建", func() {
			factory := NewRepositoryFactory(nil)

			So(factory, ShouldNotBeNil)
			So(factory.client, ShouldBeNil)
		})
	})
}

func TestRepositoryFactory_RawEvents(t *testing.T) {
	Convey("TestRepositoryFactory_RawEvents", t, func() {
		Convey("返回 RawEventRepository", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo := factory.RawEvents()

			So(repo, ShouldNotBeNil)
		})

		Convey("延迟初始化 - 多次调用返回相同实例", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo1 := factory.RawEvents()
			repo2 := factory.RawEvents()

			So(repo1, ShouldEqual, repo2)
		})
	})
}

func TestRepositoryFactory_FaultPoints(t *testing.T) {
	Convey("TestRepositoryFactory_FaultPoints", t, func() {
		Convey("返回 FaultPointRepository", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo := factory.FaultPoints()

			So(repo, ShouldNotBeNil)
		})

		Convey("延迟初始化 - 多次调用返回相同实例", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo1 := factory.FaultPoints()
			repo2 := factory.FaultPoints()

			So(repo1, ShouldEqual, repo2)
		})
	})
}

func TestRepositoryFactory_FaultPointRelations(t *testing.T) {
	Convey("TestRepositoryFactory_FaultPointRelations", t, func() {
		Convey("返回 FaultPointRelationRepository", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo := factory.FaultPointRelations()

			So(repo, ShouldNotBeNil)
		})

		Convey("延迟初始化 - 多次调用返回相同实例", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo1 := factory.FaultPointRelations()
			repo2 := factory.FaultPointRelations()

			So(repo1, ShouldEqual, repo2)
		})
	})
}

func TestRepositoryFactory_Problems(t *testing.T) {
	Convey("TestRepositoryFactory_Problems", t, func() {
		Convey("返回 ProblemRepository", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo := factory.Problems()

			So(repo, ShouldNotBeNil)
		})

		Convey("延迟初始化 - 多次调用返回相同实例", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo1 := factory.Problems()
			repo2 := factory.Problems()

			So(repo1, ShouldEqual, repo2)
		})
	})
}

func TestRepositoryFactory_FaultCausals(t *testing.T) {
	Convey("TestRepositoryFactory_FaultCausals", t, func() {
		Convey("返回 FaultCausalRepository", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo := factory.FaultCausals()

			So(repo, ShouldNotBeNil)
		})

		Convey("延迟初始化 - 多次调用返回相同实例", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo1 := factory.FaultCausals()
			repo2 := factory.FaultCausals()

			So(repo1, ShouldEqual, repo2)
		})
	})
}

func TestRepositoryFactory_FaultCausalRelations(t *testing.T) {
	Convey("TestRepositoryFactory_FaultCausalRelations", t, func() {
		Convey("返回 FaultCausalRelationRepository", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo := factory.FaultCausalRelations()

			So(repo, ShouldNotBeNil)
		})

		Convey("延迟初始化 - 多次调用返回相同实例", func() {
			client := newMockClient(200, `{}`)
			factory := NewRepositoryFactory(client)

			repo1 := factory.FaultCausalRelations()
			repo2 := factory.FaultCausalRelations()

			So(repo1, ShouldEqual, repo2)
		})
	})
}
