package models

//CharacterMeta -
type CharacterMeta struct {
	Character     Character `json:"character,omitempty"`
	EquippedItems []Item    `json:"equipped_items,omitempty"`
}

//CharacterMedia -
type CharacterMedia struct {
	Character Character `json:"character,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	BustURL   string    `json:"bust_url,omitempty"`
	RenderURL string    `json:"render_url,omitempty"`
}

//CharacterAppearance -
type CharacterAppearance struct {
	Character     Character   `json:"character,omitempty"`
	PlayableRace  interface{} `json:"playable_race,omitempty"`
	PlayableClass interface{} `json:"playable_class,omitempty"`
	ActiveSpec    interface{} `json:"active_spec,omitempty"`
	Gender        interface{} `json:"gender,omitempty"`
	Faction       interface{} `json:"faction,omitempty"`
	GuildCrest    interface{} `json:"guild_crest,omitempty"`
	Apppearance   interface{} `json:"apppearance,omitempty"`
	Items         []Item      `json:"items,omitempty"`
}

//Character -
type Character struct {
	ID int `json:"id,omitempty"`
	//Key   string `json:"key,omitempty"`
	Name  string `json:"name,omitempty"`
	Realm Realm  `json:"realm,omitempty"`
}

//Realm -
type Realm struct {
	ID   int               `json:"id,omitempty"`
	Name map[string]string `json:"name,omitempty"`
	Slug string            `json:"slug,omitempty"`
}

type Item struct {
	ID            int               `json:"id,omitempty"`
	Item          ItemKey           `json:"item,omitempty"`
	Name          map[string]string `json:"name,omitempty"`
	Description   map[string]string `json:"description,omitempty"`
	Slot          ItemSlot          `json:"slot,omitempty"`
	Quantity      interface{}       `json:"quantity,omitempty"`
	Context       interface{}       `json:"context,omitempty"`
	Quality       ItemSlot          `json:"quality,omitempty"`
	ItemClass     ItemClass         `json:"item_class,omitempty"`
	ItemSubClass  ItemClass         `json:"item_sub_class,omitempty"`
	InventoryType ItemSlot          `json:"inventory_type,omitempty"`
	Binding       ItemSlot          `json:"binding,omitempty"`
	Armor         ItemStats         `json:"armor,omitempty"`
	Media         ItemKey           `json:"media,omitempty"`

	Stats   []ItemStats   `json:"stats,omitempty"`
	Spells  []interface{} `json:"spells,omitempty"`
	Sockets []interface{} `json:"sockets,omitempty"`

	Level         interface{} `json:"level,omitempty"`
	RequiredLevel interface{} `json:"required_level,omitempty"`
	PurchasePrice interface{} `json:"purchase_price,omitempty"`
	SellPrice     interface{} `json:"sell_price,omitempty"`
	MaxCount      interface{} `json:"max_count,omitempty"`

	//Appearance
	Enchant                  interface{} `json:"enchant,omitempty"`
	ItemAppearanceModifierID interface{} `json:"item_appearance_modifier_id,omitempty"`
	InternalSlotID           interface{} `json:"internal_slot_id,omitempty"`
	Subclass                 interface{} `json:"subclass,omitempty"`
}

type ItemKey struct {
	ID  int               `json:"id,omitempty"`
	Key map[string]string `json:"key,omitempty"`
}
type ItemMedia struct {
	ID     int         `json:"id,omitempty"`
	Assets interface{} `json:"assets,omitempty"`
}

type ItemClass struct {
	//Key  string            `json:"key,omitempty"`
	Name map[string]string `json:"name,omitempty"`
	ID   int               `json:"id,omitempty"`
}

type ItemSlot struct {
	Type                 string            `json:"type,omitempty"`
	Name                 map[string]string `json:"name,omitempty"`
	ModifiedAppearanceID int               `json:"modified_appearance_id,omitempty"`
}

type ItemStats struct {
	Type    map[string]interface{} `json:"type,omitempty"`
	Value   int                    `json:"value,omitempty"`
	Display interface{}            `json:"display,omitempty"`
}
