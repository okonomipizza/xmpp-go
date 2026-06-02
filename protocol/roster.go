package protocol

import (
	"errors"
	"fmt"

	"github.com/okonomipizza/xmpp-go/jid"
)

const rosterNS = "jabber:iq:roster"

// ErrRosterPushIgnored は RFC 6121 Section 2.1.6 に従い無視すべき roster push である。
var ErrRosterPushIgnored = errors.New("protocol: roster push ignored")

// RosterItem は roster の 1 連絡先を表す (RFC 6121 Section 2)。
type RosterItem struct {
	JID          jid.JID
	Name         string
	Subscription string
	Ask          string
	Groups       []string
}

// RosterSetItem は roster set の 1 <item/> 用データ (RFC 6121 Section 2.1.5 / 2.1.6)。
type RosterSetItem struct {
	JID          jid.JID
	Name         string
	Subscription string
	Groups       []string
}

// RosterQueryGetBytes は roster get 用の <query/> 子要素を返す。
func RosterQueryGetBytes() []byte {
	return []byte("<query xmlns='" + rosterNS + "'/>")
}

// RosterIQGetBytes は RFC 6121 Example 1 の roster get IQ を返す。
func RosterIQGetBytes(from jid.JID, id string) ([]byte, error) {
	return IQStanzaBytes(from, jid.JID{}, id, "get", RosterQueryGetBytes())
}

// RosterSetItemBytes は roster set 用の <item/> 要素を返す。
// subscription が "remove" のときは jid のみ (RFC 6121 Section 2.1.6)。
func RosterSetItemBytes(item RosterSetItem) ([]byte, error) {
	if err := validateRosterItemJID(item.JID); err != nil {
		return nil, err
	}
	jidStr := item.JID.String()
	if err := validateStanzaAttr("jid", jidStr); err != nil {
		return nil, err
	}
	sub := item.Subscription
	if err := validateRosterSetSubscription(sub); err != nil {
		return nil, err
	}
	if sub == "remove" {
		if item.Name != "" || len(item.Groups) > 0 {
			return nil, errors.New("protocol: roster remove item must only contain jid and subscription")
		}
		return []byte("<item jid='" + jidStr + "' subscription='remove'/>"), nil
	}
	open := "<item jid='" + jidStr + "'"
	if item.Name != "" {
		if err := validateStanzaAttr("name", item.Name); err != nil {
			return nil, err
		}
		open += " name='" + item.Name + "'"
	}
	if sub != "" {
		if err := validateStanzaAttr("subscription", sub); err != nil {
			return nil, err
		}
		open += " subscription='" + sub + "'"
	}
	if err := validateRosterSetGroups(item.Groups); err != nil {
		return nil, err
	}
	if len(item.Groups) == 0 {
		return []byte(open + "/>"), nil
	}
	var inner []byte
	for _, g := range item.Groups {
		inner = append(inner, "<group>"...)
		inner = append(inner, escapeXMLText(g)...)
		inner = append(inner, "</group>"...)
	}
	out := append([]byte(open+">"), inner...)
	out = append(out, "</item>"...)
	return out, nil
}

// RosterQuerySetBytes は roster set 用の <query/> 子要素を返す (item は 1 件のみ、RFC 6121 Section 2.1.5)。
func RosterQuerySetBytes(item RosterSetItem) ([]byte, error) {
	itemXML, err := RosterSetItemBytes(item)
	if err != nil {
		return nil, err
	}
	var out []byte
	out = append(out, "<query xmlns='"+rosterNS+"'>"...)
	out = append(out, itemXML...)
	out = append(out, "</query>"...)
	return out, nil
}

// RosterIQSetBytes は roster set IQ を返す (RFC 6121 Section 2.3.1 等)。
func RosterIQSetBytes(from jid.JID, id string, item RosterSetItem) ([]byte, error) {
	query, err := RosterQuerySetBytes(item)
	if err != nil {
		return nil, err
	}
	return IQStanzaBytes(from, jid.JID{}, id, "set", query)
}

// ParseRosterPush は roster push (iq type='set') から 1 件の item を返す (RFC 6121 Section 2.1.6)。
// from が空、または accountBare と完全一致しない push は ErrRosterPushIgnored を返す。
func ParseRosterPush(ev IQEvent, accountBare jid.JID) (RosterItem, error) {
	if ev.Type != "set" {
		return RosterItem{}, fmt.Errorf("protocol: roster push requires iq type set, got %q", ev.Type)
	}
	if accountBare.IsEmpty() || !accountBare.IsBare() {
		return RosterItem{}, errors.New("protocol: account bare jid required")
	}
	if !ev.From.IsEmpty() && !ev.From.Equal(accountBare) {
		return RosterItem{}, ErrRosterPushIgnored
	}
	items, err := parseRosterPushItems(ev.Payload)
	if err != nil {
		return RosterItem{}, err
	}
	if len(items) != 1 {
		return RosterItem{}, fmt.Errorf("protocol: roster push requires exactly one item, got %d", len(items))
	}
	return items[0], nil
}

func validateRosterItemJID(j jid.JID) error {
	if j.IsEmpty() {
		return errors.New("protocol: roster item jid required")
	}
	if j.IsFull() {
		return errors.New("protocol: roster item jid must be bare")
	}
	return nil
}

func validateRosterSetSubscription(sub string) error {
	switch sub {
	case "", "remove":
		return nil
	default:
		return fmt.Errorf("protocol: invalid roster set subscription %q", sub)
	}
}

func validateRosterSetGroups(groups []string) error {
	seen := make(map[string]struct{}, len(groups))
	for _, g := range groups {
		if g == "" {
			return errors.New("protocol: roster set group must not be empty")
		}
		if _, dup := seen[g]; dup {
			return fmt.Errorf("protocol: duplicate roster set group %q", g)
		}
		seen[g] = struct{}{}
	}
	return nil
}

// ParseRosterItems は IQ の Payload から roster item を抽出する。
// Payload は <query xmlns='jabber:iq:roster'> 要素、またはその親要素の inner でもよい。
func ParseRosterItems(iqPayload []byte) ([]RosterItem, error) {
	if isRosterQueryElement(iqPayload) {
		_, selfClosing, ok := findOpenTagEnd(iqPayload, 0)
		if !ok {
			return nil, errors.New("protocol: invalid roster query")
		}
		if selfClosing {
			return nil, nil
		}
		return parseRosterQueryItems(iqPayload, false)
	}
	for _, ch := range parseDirectChildElements(iqPayload) {
		if ch.name == "query" && elementAttrValue(ch.raw, "xmlns") == rosterNS {
			return parseRosterQueryItems(ch.raw, false)
		}
	}
	return nil, errors.New("protocol: roster query not found")
}

func isRosterQueryElement(raw []byte) bool {
	if len(raw) == 0 || raw[0] != '<' {
		return false
	}
	openEnd, _, ok := findOpenTagEnd(raw, 0)
	if !ok {
		return false
	}
	return openTagLocalName(raw[:openEnd+1]) == "query" && elementAttrValue(raw, "xmlns") == rosterNS
}

func parseRosterPushItems(payload []byte) ([]RosterItem, error) {
	if isRosterQueryElement(payload) {
		_, selfClosing, ok := findOpenTagEnd(payload, 0)
		if !ok {
			return nil, errors.New("protocol: invalid roster query")
		}
		if selfClosing {
			return nil, nil
		}
		return parseRosterQueryItems(payload, true)
	}
	for _, ch := range parseDirectChildElements(payload) {
		if ch.name == "query" && elementAttrValue(ch.raw, "xmlns") == rosterNS {
			return parseRosterQueryItems(ch.raw, true)
		}
	}
	return nil, errors.New("protocol: roster query not found")
}

func parseRosterQueryItems(queryRaw []byte, push bool) ([]RosterItem, error) {
	var items []RosterItem
	for _, ch := range parseDirectChildElements(queryRaw) {
		if ch.name != "item" {
			continue
		}
		item, err := parseRosterItem(ch.raw, push)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func parseRosterItem(itemRaw []byte, push bool) (RosterItem, error) {
	jidStr := elementAttrValue(itemRaw, "jid")
	if jidStr == "" {
		return RosterItem{}, errors.New("protocol: roster item missing jid")
	}
	j, err := jid.Parse(jidStr)
	if err != nil {
		return RosterItem{}, err
	}
	var groups []string
	for _, ch := range parseDirectChildElements(itemRaw) {
		if ch.name == "group" {
			text := unescapeXMLText(elementTextContent(ch.raw))
			if text != "" {
				groups = append(groups, text)
			}
		}
	}
	sub := elementAttrValue(itemRaw, "subscription")
	if push {
		sub = normalizeRosterPushSubscription(sub)
	} else {
		sub = normalizeRosterResultSubscription(sub)
	}
	return RosterItem{
		JID:          j,
		Name:         elementAttrValue(itemRaw, "name"),
		Subscription: sub,
		Ask:          elementAttrValue(itemRaw, "ask"),
		Groups:       groups,
	}, nil
}

func normalizeRosterResultSubscription(sub string) string {
	switch sub {
	case "none", "to", "from", "both":
		return sub
	default:
		return ""
	}
}

func normalizeRosterPushSubscription(sub string) string {
	switch sub {
	case "none", "to", "from", "both", "remove":
		return sub
	default:
		return ""
	}
}
