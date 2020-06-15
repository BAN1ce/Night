package sub

import (
	"live/pkg/utils"
	"strings"
)

func GetSub(topic string) ([]*Node, []string) {
	topicSlice := strings.Split(topic, "/")
	nodes := GetWildCards(topicSlice)
	clients := GetHashSub(topic)
	return nodes, clients
}

func Sub(topic, clientIdentifier string, topicSlice []string) {

	if utils.HasWildcard(topicSlice) {
		AddTreeSub(topicSlice, clientIdentifier)
	} else {
		AddHashSub(topic, clientIdentifier)
	}

}
