package messageIdStorage

var ch = make(map[int64]chan int)

func Set(chatID int64, messageID int) {
	_, exists := ch[chatID]
	if exists {
		ch[chatID] <- messageID
	} else {
		ch[chatID] = make(chan int, 1000)
		ch[chatID] <- messageID
	}

}

func Get(chatID int64) []int {
	_, exists := ch[chatID]
	var messageIDs []int
	if exists {
		close(ch[chatID])
		for i := range ch[chatID] {
			messageIDs = append(messageIDs, i)
		}
		delete(ch, chatID)
	}
	return messageIDs
}
