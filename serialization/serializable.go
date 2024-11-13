package serialization

type Serializable interface {
	Serialize()
	Deserialize()
}