package async_op_apis

import (
	"mmo/cmd/async_op/db_model"
	"mmo/ginm/source/async_op"
	"mmo/ginm/source/inter"
)

func AsyncUserSaveData(request inter.Request) *async_op.AsyncOpResult {

	opId := 1 // player's unique identifier Id (玩家的唯一标识Id)
	asyncResult := async_op.NewAsyncOpResult(request.GetConn(), request.GetWorkerID())

	async_op.Process(
		int(opId),
		func() {
			// perform db operation (执行db操作)
			user := db_model.SaveUserData()

			// set async return result (设置异步返回结果)
			asyncResult.SetAsyncOpResult(user)

			// test active exception (测试主动异常)
			/*
				a := 0
				b := 1
				c := b / a
				fmt.Println(c)
			*/
		},
	)

	return asyncResult
}
