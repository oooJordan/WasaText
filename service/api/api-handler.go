package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	//LOGIN UTENTE
	rt.router.POST("/session", rt.wrap(rt.loginUser))
	rt.router.GET("/users", rt.wrap(rt.listUsers))
	rt.router.PUT("/users/:user_id/username", rt.wrap(rt.updateUsername))

	// Register routes
	/*

		rt.router.POST("/session/profile_image", rt.wrap(rt.updateProfileImage))
		rt.router.GET("/session/profile_image", rt.wrap(rt.getProfileImage))
		rt.router.POST("/upload", rt.wrap(rt.uploadImage))


		rt.router.GET("/users", rt.wrap(rt.listUsers))

		rt.router.GET("/conversation", rt.wrap(rt.conversationsList))
		rt.router.GET("/conversation", rt.wrap(rt.getMyConversations))
		rt.router.POST("/conversation", rt.wrap(rt.newConversation))
		rt.router.GET("/conversation/:conversation_id", rt.wrap(rt.messageHistory))
		rt.router.POST("/conversation/:conversation_id", rt.wrap(rt.sendNewMessage))

		rt.router.POST("/conversation/:conversation_id/messages/:message_id", rt.wrap(rt.forwardMessage))
		rt.router.DELETE("/conversation/:conversation_id/messages/:message_id", rt.wrap(rt.deleteMessage))
		rt.router.PATCH("/conversation/:conversation_id/messages/:message_id/comment", rt.wrap(rt.commentMessage))
		rt.router.DELETE("/conversation/:conversation_id/messages/:message_id/comment/:comment_id", rt.wrap(rt.removeReaction))

		rt.router.PATCH("/conversation/:conversation_id/names", rt.wrap(rt.addUser))
		rt.router.POST("/conversation/:conversation_id/leave", rt.wrap(rt.leaveGroup))
		rt.router.PUT("/conversation/:conversation_id/name", rt.wrap(rt.renameGroup))
		rt.router.PUT("/conversation/:conversation_id/groupimage", rt.wrap(rt.updateGroupImage))
	*/
	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
