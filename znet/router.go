package znet

import "github.com/zuishabi/zinx/ziface"

// BaseRouter 实现router时，先嵌入这个BaseRouter基类，然后根据需要对这个基类的方法进行重写就好了
type BaseRouter struct{}

//这里之所以BaseRouter的方法都为空，是因为有的Router不希望有PreHandle、PostHandle这两个业务
//所以Router全部继承BaseRouter的好处就是不需要实现PreHandle，PostHandle

// PreHandle 在处理conn业务之前的钩子方法Hook
func (br *BaseRouter) PreHandle(request ziface.IRequest) {}

// Handle 在处理conn业务的主方法hook
func (br *BaseRouter) Handle(request ziface.IRequest) {}

// PostHandle 在处理conn业务之后的钩子方法hook
func (br *BaseRouter) PostHandle(request ziface.IRequest) {}
