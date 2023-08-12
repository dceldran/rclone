package api

type InputPeer struct {
  _     'inputPeer'
  user_id     number
  access_hash     string
}

type InputUser struct {
  _     'inputUser'
  user_id     number
  access_hash     string
}

type InputChannel struct {
  _     'inputChannel'
  channel_id     number
  access_hash     string
}

type InputMessageID struct {
  _     'inputMessageID'
  id     number
}

type MessagesGetMessages struct {
  _     'messages.getMessages'
  id     InputMessageID[]
}

type MessagesGetMessagesResponse struct {
  messages     Message[]
}

type Message struct {
  _     'message'
  id     number
  media     Media
}

type Media struct {
  _     'messageMediaDocument'
  document     Document
}

type Document struct {
  _     'document'
  id     string
  access_hash     string
  size     number
  mime_type     string
  file_reference     Uint8Array
}