package api

import (
	"net/http"
	"os"
	"path/filepath"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	rt.router.POST("/session", rt.wrap(rt.loginUser))
	rt.router.GET("/users", rt.wrap(rt.listUsers))
	rt.router.PUT("/username", rt.wrap(rt.UpdateUsername))
	rt.router.POST("/conversations", rt.wrap(rt.newConversation))
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.PUT("/profile_image", rt.wrap(rt.updateProfileImage))
	rt.router.GET("/profile_image", rt.wrap(rt.getProfileImage))
	rt.router.PUT("/conversation/:conversation_id/names", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/conversation/:conversation_id/membership", rt.wrap(rt.leaveGroup))
	rt.router.PUT("/conversation/:conversation_id/name", rt.wrap(rt.renameGroup))
	rt.router.PUT("/conversation/:conversation_id/groupimage", rt.wrap(rt.updateGroupImage))
	rt.router.POST("/conversation/:conversation_id", rt.wrap(rt.sendNewMessage))
	rt.router.GET("/conversation/:conversation_id", rt.wrap(rt.messageHistory))
	rt.router.POST("/conversation/:conversation_id/messages/:message_id", rt.wrap(rt.forwardMessage))
	rt.router.DELETE("/conversation/:conversation_id/messages/:message_id", rt.wrap(rt.deleteMessage))
	rt.router.PUT("/conversation/:conversation_id/messages/:message_id/comment", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/conversation/:conversation_id/messages/:message_id/comment", rt.wrap(rt.removeReaction))
	rt.router.POST("/upload", rt.wrap(rt.uploadImage))
	rt.router.ServeFiles("/uploads/*filepath", http.Dir("./uploads"))
	cwd, _ := os.Getwd()
	rt.router.ServeFiles("/defaultimage/*filepath", http.Dir(filepath.Join(cwd, "defaultimage")))
	// Register routes

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
