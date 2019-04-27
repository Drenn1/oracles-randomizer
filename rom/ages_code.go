package rom

import (
	"strings"
)

func newAgesRomBanks() *romBanks {
	asm, err := newAssembler()
	if err != nil {
		panic(err)
	}

	r := romBanks{
		endOfBank: make([]uint16, 0x40),
		assembler: asm,
	}

	r.endOfBank[0x00] = 0x3ef8
	r.endOfBank[0x01] = 0x7fc3
	r.endOfBank[0x02] = 0x7e93
	r.endOfBank[0x03] = 0x7ebd
	r.endOfBank[0x04] = 0x7edb
	r.endOfBank[0x05] = 0x7d9d
	r.endOfBank[0x06] = 0x7a31
	r.endOfBank[0x08] = 0x7f60
	r.endOfBank[0x09] = 0x7dee
	r.endOfBank[0x0a] = 0x7e09
	r.endOfBank[0x0b] = 0x7fa8
	r.endOfBank[0x0c] = 0x7f94
	r.endOfBank[0x0f] = 0x7f90
	r.endOfBank[0x10] = 0x7ef4
	r.endOfBank[0x11] = 0x7f73
	r.endOfBank[0x12] = 0x7e8f
	r.endOfBank[0x15] = 0x7bfb
	r.endOfBank[0x16] = 0x7e03
	r.endOfBank[0x38] = 0x6b00 // to be safe
	r.endOfBank[0x3f] = 0x7d0a

	// do this before loading asm files, since the size of this table varies
	// with the number of checks.
	r.appendToBank(0x06, "collectModeTable", makeAgesCollectModeTable())

	r.applyAsmFiles([]string{"/asm/common.yaml", "/asm/ages.yaml"})

	return &r
}

func initAgesEOB() {
	r := newAgesRomBanks()
	globalRomBanks = r

	// bank 00

	r.replaceAsm(0x00, 0x0c9a,
		"ld h,a; ld a,(ff00+b7)", "call filterMusic")
	r.replaceAsm(0x00, 0x3e56,
		"inc a; cp a,11", "call checkMakuState")

	compareRoom := addrString(r.assembler.getDef("compareRoom"))
	searchDoubleKey := addrString(r.assembler.getDef("searchDoubleKey"))
	findObjectWithId := addrString(r.assembler.getDef("findObjectWithId"))

	// bank 01

	// use a different invalid tile table for time warping if link doesn't have
	// flippers.
	noFlippersTable := r.appendToBank(0x01, "no flippers table",
		"\xf3\x00\xfe\x00\xff\x00\xe4\x00\xe5\x00\xe6\x00\xe7\x00\xe8\x00"+
			"\xe9\x00\xfc\x01\xfa\x00\xe0\x00\xe1\x00\xe2\x00\xe3\x00\x00")
	dontDrownLink := r.appendToBank(0x01, "don't drown link",
		"\x21\x17\x63\xfa\x9f\xc6\xe6\x40\xc0\x21"+noFlippersTable+"\xc9")
	r.replace(0x01, 0x6301, "call don't drown link",
		"\x21\x17\x63", "\xcd"+dontDrownLink)

	// bank 02

	r.replaceMultiple([]Addr{{0x02, 0x6133}, {0x02, 0x618b}}, "tree warp jump",
		"\xc2\xba\x4f", "\xc4"+addrString(r.assembler.getDef("treeWarp")))
	r.replaceAsm(0x02, 0x5fcb, "call setMusicVolume", "call devWarp")

	r.replaceAsm(0x02, 0x5ff9,
		"call _mapMenu_checkRoomVisited", "call checkTreeVisited")
	r.replaceAsm(0x02, 0x66a9,
		"call _mapMenu_checkRoomVisited", "call checkTreeVisited")
	r.replaceAsm(0x02, 0x619d,
		"call _mapMenu_checkCursorRoomVisited", "call checkCursorVisited")
	r.replaceAsm(0x02, 0x6245,
		"ld a,(wMapMenu_cursorIndex)", "jp displayPortalPopups")

	r.replaceAsm(0x02, 0x56dd,
		"ld a,(wInventorySubmenu1CursorPos)", "call openRingList")
	r.replaceAsm(0x02, 0x7019,
		"call _ringMenu_updateSelectedRingFromList", "call autoEquipRing")
	r.replaceAsm(0x02, 0x5074,
		"call setMusicVolume", "call ringListGfxFix")

	// bank 03

	r.replaceAsm(0x03, 0x4d6b,
		"call decHlRef16WithCap", "call skipCapcom")
	r.replaceAsm(0x03, 0x6e97,
		"jp setGlobalFlag", "jp setInitialFlags")

	// bank 04

	r.replaceAsm(0x00, 0x38c0,
		"call applyAllTileSubstitutions", "call applyExtraTileSubstitutions")

	// treat the d2 present entrance like the d2 past entrance, and reset the
	// water level when entering jabu (see logic comments).
	replaceWarpEnter := r.appendToBank(0x04, "replace warp enter",
		"\xc5\x01\x00\x83\xcd"+compareRoom+"\x20\x04\xc1\x3e\x01\xc9"+
			"\x01\x02\x90\xcd"+compareRoom+"\xc1\x20\x05\x3e\x21\xea\xe9\xc6"+
			"\xfa\x2d\xcc\xc9")
	r.replace(0x04, 0x4630, "call replace warp enter",
		"\xfa\x2d\xcc", "\xcd"+replaceWarpEnter)
	// d2: exit into the present if the past entrance is closed.
	replaceWarpExit := r.appendToBank(0x00, "replace warp exit",
		"\xea\x48\xcc\xfe\x83\xc0\xfa\x83\xc8\xe6\x80\xc0"+
			"\xfa\x47\xcc\xe6\x0f\xfe\x01\xc0"+
			"\xfa\x47\xcc\xe6\xf0\xea\x47\xcc\xc9")
	r.replace(0x04, 0x45e8, "call replace warp exit normal",
		"\xea\x48\xcc", "\xcd"+replaceWarpExit)
	r.replace(0x0a, 0x4738, "call replace warp exit essence",
		"\xea\x48\xcc", "\xcd"+replaceWarpExit)

	// bank 05

	r.replaceAsm(0x05, 0x6083,
		"call lookupCollisionTable", "call cliffLookup")

	// prevent link from surfacing from underwater without mermaid suit. this
	// is probably only relevant for the sea of no return.
	preventSurface := r.appendToBank(0x05, "prevent surface",
		"\xfa\x91\xcc\xb7\xc0\xfa\xa3\xc6\xe6\x04\xfe\x04\xc9")
	r.replace(0x05, 0x516c, "call prevent surface",
		"\xfa\x91\xcc\xb7", "\xcd"+preventSurface+"\x00")

	// bank 06

	// burning the first tree in yoll graveyard should set room flag 1 so that
	// it can be gone for good.
	removeYollTree := r.appendToBank(0x06, "remove yoll tree",
		"\xf5\xf0\x8f\xfe\x0c\x20\x0f"+
			"\xc5\x01\x00\x6b\xcd"+compareRoom+"\x20\x05"+
			"\x21\x6b\xc7\xcb\xce\xc1\xf1\x21\x26\xc6\xc9")
	r.replace(0x06, 0x47aa, "call remove yoll tree",
		"\x21\x26\xc6", "\xcd"+removeYollTree)

	// reenter a warp tile that link is standing on when playing the tune of
	// currents (useful if you warp into a patch of bushes). also activate the
	// west present crescent island portal.
	reenterCurrentsWarp := r.appendToBank(0x06, "special currents actions",
		"\xc5\x01\x00\xa9\xcd"+compareRoom+"\xc1\x20\x11"+ // island portal
			"\xd5\x3e\xe1\xcd"+findObjectWithId+"\x20\x05"+ // cont.
			"\x1e\x44\x3e\x02\x12\xd1\xc3\x08\x4e"+ // cont.
			"\xfa\x34\xcc\xf5\xd5\x3e\xde\xcd"+findObjectWithId+ // reenter
			"\x20\x05\x1e\x44\x3e\x02\x12\xd1\xf1\xc3\x37\x4e") // cont.
	r.replace(0x06, 0x4e34, "call special currents actions",
		"\xfa\x34\xcc", "\xc3"+reenterCurrentsWarp)

	// set text index for portal sign on crescent island.
	setPortalSignText := r.appendToBank(0x06, "set portal sign text",
		"\x01\x00\xa9\xcd"+compareRoom+"\x01\x01\x09\xc0\x01\x01\x56\xc9")
	r.replace(0x06, 0x40e7, "call set portal sign text",
		"\x01\x01\x09", "\xcd"+setPortalSignText)

	// use expert's or fist ring with only one button unequipped.
	r.replaceAsm(0x06, 0x4969, "ret nz", "nop")

	// bank 16 (pt. 1)

	getTreasureDataBCE := addrString(r.assembler.getDef("getTreasureDataBCE"))

	// bank 09

	// set treasure ID 07 (rod of seasons) when buying the 150 rupee shop item,
	// so that the shop can check this specific ID.
	shopSetFakeID := r.appendToBank(0x09, "shop set fake ID",
		"\xfe\x0d\x20\x05\x21\x9a\xc6\xcb\xfe\x21\xf7\x44\xc9")
	r.replace(0x09, 0x4418, "call shop set fake ID",
		"\x21\xf7\x44", "\xcd"+shopSetFakeID)

	// set treasure ID 08 (magnet gloves) when getting item from south shore
	// dirt pile.
	digSetFakeID := r.appendToBank(0x09, "dirt set fake ID",
		"\xc5\x01\x00\x98\xcd"+compareRoom+"\xc1\xc0\xe5\x21\x9b\xc6\xcb\xc6"+
			"\xe1\xc9")
	// set treasure ID 13 (slingshot) when getting first item from tingle.
	tingleSetFakeID := r.appendToBank(0x09, "tingle set fake ID",
		"\xc5\x01\x00\x79\xcd"+compareRoom+"\xc1\xc0\xe5\x21\x9c\xc6\xcb\xde"+
			"\xe1\xc9")
	// set treasure ID 1e (fool's ore) for symmetry city brother.
	brotherSetFakeID := r.appendToBank(0x09, "brother set fake ID",
		"\xc5\x01\x03\x6e\xcd"+compareRoom+"\x28\x04\x04\xcd"+compareRoom+
			"\xc1\xc0\xe5\x21\x9d\xc6\xcb\xf6\xe1\xc9")
	// set treasure ID 10 (nothing) for king zora.
	kingZoraSetFakeID := r.appendToBank(0x09, "king zora set fake ID",
		"\xc5\x01\x05\xab\xcd"+compareRoom+"\xc1\xc0\xe5\x21\x9c\xc6\xcb\xc6"+
			"\xe1\xc9")
	// set treasure ID 12 (nothing) for first goron dance, and 14 (nothing) for
	// the second. if you're in the present, it's always 12. if you're in the
	// past, it's 12 iff you don't have letter of introduction.
	goronDanceSetFakeID := r.appendToBank(0x09, "dance 1 set fake ID",
		"\xc5\x01\x02\xed\xcd"+compareRoom+"\xc1\x28\x12"+ // present
			"\xc5\x01\x02\xef\xcd"+compareRoom+"\xc1\xc0"+ // past
			"\x3e\x59\xcd\x48\x17\x3e\x10\x38\x02\x3e\x04"+
			"\xe5\x21\x9c\xc6\xb6\x77\xe1\xc9")
	// set flag for d6 past and present boss keys whether you get the key in
	// past or present.
	setD6BossKey := r.appendToBank(0x09, "set d6 boss key",
		"\x7b\xfe\x31\xc0\xfa\x39\xcc\xfe\x06\x28\x03\xfe\x0c\xc0"+
			"\xe5\x21\x82\xc6\xcb\xf6\x23\xcb\xe6\xe1\xc9")
	// refill all seeds when picking up a seed satchel.
	refillSeedSatchel := r.appendToBank(0x09, "refill seed satchel",
		"\x7b\xfe\x19\xc0"+
			"\xc5\xd5\xe5\x21\xb4\xc6\x34\xcd\x0c\x18\x35\xe1\xd1\xc1\xc9")
	// give 20 seeds when picking up the seed shooter.
	fillSeedShooter := r.appendToBank(0x09, "fill seed shooter",
		"\x7b\xfe\x0f\xc0\xc5\x3e\x20\x0e\x20\xcd\x1c\x17\xc1\xc9")
	// give flute the correct icon and make it functional from the start.
	activateFlute := r.appendToBank(0x09, "activate flute",
		"\x7b\xfe\x0e\xc0"+
			"\x79\xd6\x0a\xea\xb5\xc6\xe5\x26\xc6\xc6\x45\x6f\x36\xc3\xe1\xc9")
	// reset maku tree to state 02 after getting the maku seed.
	makuSeedResetState := r.appendToBank(0x09, "maku seed reset state",
		"\x7b\xfe\x36\xc0\x3e\x02\xea\xe8\xc6\xc9")
	// this function checks all the above conditions when collecting an item.
	handleGetItem := r.appendToBank(0x09, "handle get item",
		"\x5f\xcd"+digSetFakeID+"\xcd"+setD6BossKey+"\xcd"+refillSeedSatchel+
			"\xcd"+fillSeedShooter+"\xcd"+activateFlute+"\xcd"+tingleSetFakeID+
			"\xcd"+brotherSetFakeID+"\xcd"+kingZoraSetFakeID+
			"\xcd"+goronDanceSetFakeID+"\xcd"+makuSeedResetState+
			"\x7b\xc3\x1c\x17")
	r.replace(0x09, 0x4c4e, "call handle get item",
		"\xcd\x1c\x17", "\xcd"+handleGetItem)

	// remove generic "you got a ring" text for rings from shops
	r.replace(0x09, 0x4580, "obtain ring text replacement (shop) 1", "\x54", "\x00")
	r.replace(0x09, 0x458a, "obtain ring text replacement (shop) 2", "\x54", "\x00")
	r.replace(0x09, 0x458b, "obtain ring text replacement (shop) 3", "\x54", "\x00")

	// remove generic "you got a ring" text for gasha nuts
	gashaNutRingText := r.appendToBank(0x0b, "remove ring text from gasha nut",
		"\x79\xfe\x04\xc2\x72\x18\xe1\xc9")
	r.replace(0x0b, 0x45bb, "remove ring text from gasha nut caller",
		"\xc3\x72\x18", "\xc3"+gashaNutRingText)

	// don't set room's item flag if it's nayru's item on the maku tree screen,
	// since link still might not have taken the maku tree's item.
	makuTreeItemFlag := r.appendToBank(0x09, "maku tree item flag",
		"\xcd\x7d\x19\xc5\x01\x38\xc7\xcd\xd6\x01\xc1\x20\x06\xfa\x0d\xd0"+
			"\xfe\x50\xc8\xcb\xee\xc9")
	r.replace(0x09, 0x4c82, "call maku tree item flag",
		"\xcd\x7d\x19", "\xc3"+makuTreeItemFlag)

	// give correct ID and param for shop item, play sound, and load correct
	// text index into temp wram address.
	shopGiveTreasure := r.appendToBank(0x09, "shop give treasure",
		"\x47\x1a\xfe\x0d\x78\x20\x08\xcd"+getTreasureDataBCE+"\x7b\xea\x0d\xcf"+
			"\x78\xcd"+handleGetItem+"\xc2\x98\x0c\x3e\x4c\xc3\x98\x0c")
	r.replace(0x09, 0x4425, "call shop give treasure",
		"\xcd\x1c\x17", "\xcd"+shopGiveTreasure)
	// display text based on above temp wram address.
	shopShowText := r.appendToBank(0x09, "shop show text",
		"\x1a\xfe\x0d\xc2\x72\x18\xfa\x0d\xcf\x06\x00\x4f"+
			"\x79\xfe\xff\xc8\xc3\x72\x18") // text $ff is ring
	r.replace(0x09, 0x4443, "call shop show text",
		"\xc2\x72\x18", "\xc2"+shopShowText)

	// bank 0a

	// make ricky appear if you have his gloves, without giving rafton rope.
	checkRickyAppear := r.appendToBank(0x0a, "check ricky appear",
		"\xcd\xf3\x31\xc0\xfa\xa3\xc6\xcb\x47\xc0\xfa\x46\xc6\xb7\xc9")
	r.replace(0x0a, 0x4bb8, "call check ricky appear",
		"\xcd\xf3\x31", "\xcd"+checkRickyAppear)

	// require giving rafton rope, even if you have the island chart.
	checkRaftonRope := r.appendToBank(0x0a, "check rafton rope",
		"\xcd\x48\x17\xd0\x3e\x15\xcd\xf3\x31\xc8\x37\xc9")
	r.replace(0x0a, 0x4d5f, "call check rafton rope",
		"\xcd\x48\x17", "\xcd"+checkRaftonRope)

	// set sub ID for south shore dig item.
	dirtSpawnItem := r.appendToBank(0x0a, "dirt spawn item",
		"\xcd\xd4\x27\xc0\xcd\x42\x22\xaf\xc9")
	r.replace(0x0a, 0x5e3e, "call dirt spawn item",
		"\xcd\xc5\x24", "\xcd"+dirtSpawnItem)

	// automatically save maku tree when saving nayru.
	saveMakuTreeWithNayru := r.appendToBank(0x0a, "save maku tree with nayru",
		"\xcd\xf9\x31\xfa\xe8\xc6\xfe\x0e\x28\x02\x3e\x02\x3d\xea\xe8\xc6"+
			"\x3e\x0c\xcd\xf9\x31\x3e\x12\xcd\xf9\x31\x3e\x3f\xcd\xf9\x31"+
			"\xe5\x21\x38\xc7\xcb\x86\x24\xcb\xfe\x2e\x48\xcb\xc6\xe1\xc9")
	r.replace(0x0a, 0x5541, "call save maku tree with nayru",
		"\xcd\xf9\x31", "\xcd"+saveMakuTreeWithNayru)

	// use a non-cutscene screen transition for exiting a dungeon via essence,
	// so that overworld music plays, and set maku tree state.
	essenceWarp := r.appendToBank(0x0a, "essence warp",
		"\x3e\x81\xea\x4b\xcc\xc3\x53\x3e")
	r.replace(0x0a, 0x4745, "call essence warp",
		"\xea\x4b\xcc", "\xcd"+essenceWarp)

	// on left side of house, swap rafton 00 (builds raft) with rafton 01 (does
	// trade sequence) if the player enters with the magic oar *and* global
	// flag 26 (rafton has built raft) is not set.
	setRaftonSubID := r.appendToBank(0x0a, "set rafton sub ID",
		"\xcd\xf3\x31\xc2\x05\x3b\xfa\xc0\xc6\xfe\x09\xc2\x5b\x4d"+
			"\x3e\x01\x12\xc3\xac\x4d")
	r.replace(0x0a, 0x4d55, "jump set rafton sub ID",
		"\xcd\xf3\x31", "\xc3"+setRaftonSubID)

	// bank 0b

	// always get item from king zora before permission to enter jabu-jabu.
	kingZoraCheck := r.appendToBank(0x0b, "king zora check",
		"\xcd\xf3\x31\xc8\x3e\x10\xcd\x48\x17\x3e\x00\xd0\x3c\xc9")
	r.replace(0x0b, 0x5464, "call king zora check",
		"\xcd\xf3\x31", "\xcd"+kingZoraCheck)

	// fairy queen cutscene: just fade back in after the fairy leaves the
	// screen, and play the long "puzzle solved" sound.
	fairyQueenFunc := r.appendToBank(0x0b, "fairy queen func",
		"\xcd\x99\x32\xaf\xea\x02\xcc\xea\x8a\xcc\x3e\x5b\xcd\x98\x0c"+
			"\x3e\x30\xcd\xf9\x31\xc9")
	r.replace(0x0b, 0x7954, "call fairy queen func",
		"\xea\x04\xcc", "\xcd"+fairyQueenFunc)

	// check either zora guard's flag for the two in sea of storms, so that
	// either can be accessed after losing the zora scale in a linked game.
	checkZoraGuards := r.appendToBank(0x0b, "check zora guards",
		"\xfa\xd7\xc7\xc5\x47\xfa\xd6\xc8\xb0\xc1\xc9")
	r.replace(0x0b, 0x61d7, "call check zora guards",
		"\xcd\x7d\x19", "\xcd"+checkZoraGuards)

	// bank 0c

	// this will be overwritten after randomization
	smallKeyDrops := r.appendToBank(0x38, "small key drops",
		makeKeyDropTable())
	lookUpKeyDropBank38 := r.appendToBank(0x38, "look up key drop bank 38",
		"\xc5\xfa\x2d\xcc\x47\xfa\x30\xcc\x4f\x21"+smallKeyDrops+ // load group/room
			"\x1e\x02\xcd"+searchDoubleKey+"\xc1\xd0\x46\x23\x4e\xc9")
	// ages has different key drop code across three different banks because
	// it's a jerk
	callBank38Code := "\xd5\xe5\x1e\x38\x21" + lookUpKeyDropBank38 +
		"\xcd\x8a\x00\xe1\xd1\xc9"
	lookUpKeyDropBank0C := r.appendToBank(0x0c, "look up key drop bank 0c",
		"\x36\x60\x2c"+callBank38Code)
	r.replace(0x0c, 0x442e, "call look up key drop bank 0c",
		"\x36\x60\x2c", "\xcd"+lookUpKeyDropBank0C)
	lookUpKeyDropBank0A := r.appendToBank(0x0a, "look up key drop bank 0A",
		"\x01\x01\x30"+callBank38Code)
	r.replace(0x0a, 0x7075, "call look up key drop bank 0A",
		"\x01\x01\x30", "\xcd"+lookUpKeyDropBank0A)
	lookUpKeyDropBank08 := r.appendToBank(0x08, "look up key drop bank 08",
		"\x01\x01\x30"+callBank38Code)
	r.replace(0x08, 0x5087, "call look up key drop bank 08",
		"\x01\x01\x30", "\xcd"+lookUpKeyDropBank08)

	// use custom script for soldier in deku forest with sub ID 0; they should
	// give an item in exchange for mystery seeds.
	soldierScriptAfter := r.appendToBank(0x0c, "soldier script after item",
		"\x97\x59\x08\x00")
	soldierScriptGive := r.appendToBank(0x0c, "soldier script give item",
		"\xeb\x9e\x98\x59\x0b\xb4\xbd\x00\x92\xe9\xcb\x02\xde\x00\x00\xb1\x20"+
			"\xc4"+soldierScriptAfter)
	soldierScriptCheck := r.appendToBank(0x0c, "soldier script check count",
		"\xb3\xbd\xff"+soldierScriptGive+"\x5d\xee")
	soldierScript := r.appendToBank(0x0c, "soldier script",
		"\xb0\x20"+soldierScriptAfter+"\xdf\x24"+soldierScriptCheck+"\x5d\xee")
	r.replace(0x09, 0x5207, "soldier script pointer", "\xee\x5d", soldierScript)

	// set room flags for other side of symmetry city bridge at end of building
	// cutscene.
	setBridgeFlag := r.appendToBank(0x15, "set bridge flag",
		"\xe5\xaf\xea\x8a\xcc\x3e\x25\xcd\xf9\x31"+
			"\x21\x24\xc7\xcb\xce\xe1\xc9")
	r.replace(0x0c, 0x7a6f, "call set bridge flag",
		"\xb9\xb6\x25", "\xe0"+setBridgeFlag)

	// skip forced ring appraisal and ring list with vasu (prevents softlock)
	r.replace(0x0c, 0x4a27, "skip vasu ring appraisal",
		"\x98\x33", "\x4a\x35")

	// bank 0f

	// set room flag for tunnel behind keep when defeating great moblin.
	setTunnelFlag := r.appendToBank(0x0f, "set tunnel flag",
		"\x21\x09\xc7\xcb\xc6\x21\xda\xca\xc9")
	r.replace(0x0f, 0x7f3e, "call set tunnel flag",
		"\x21\x09\xc7", "\xcd"+setTunnelFlag)

	// bank 10

	// keep black tower in initial state until the player got the item from the
	// worker.
	blackTowerCheck := r.appendToBank(0x10, "black tower check",
		"\x21\x27\x79\xc8\xfa\xe1\xc9\xe6\x20\xc9")
	r.replace(0x10, 0x7914, "call black tower check",
		"\x21\x27\x79", "\xcd"+blackTowerCheck)

	// don't let echoes activate the special crescent island portal.
	echoesPortalCheck := r.appendToBank(0x10, "echoes portal check",
		"\xc5\x01\x00\xa9\xcd"+compareRoom+"\xc1\xfa\x8d\xcc\xc0\x3d\xc9")
	r.replace(0x10, 0x7d88, "call echoes portal check",
		"\xfa\x8d\xcc", "\xcd"+echoesPortalCheck)

	// bank 11

	// allow collection of seeds with only shooter and no satchel
	checkSeedHarvest := r.appendToBank(0x11, "check seed harvest",
		"\xcd\x48\x17\xd8\x3e\x0f\xc3\x48\x17")
	r.replace(0x11, 0x4aba, "call check seed harvest",
		"\xcd\x48\x17", "\xcd"+checkSeedHarvest)

	// bank 12

	// add time portal interaction in symmetry city past, to avoid softlock if
	// player only has echoes.
	symmetryPastPortal := r.appendToBank(0x12, "symmetry past portal",
		"\xf1\xdc\x05\xf2\xe1\x00\x68\x18\xfe")
	r.replace(0x12, 0x5e91, "symmetry past portal pointer",
		"\xf1\xdc\x05", "\xf3"+symmetryPastPortal)
	// add one to nuun highlands too.
	nuunPortalOtherObjects := r.appendToBank(0x12, "nuun portal other objects",
		"\xf2\x9a\x00\x68\x48\x9a\x01\x58\x58\x9a\x02\x58\x68\x9a\x03\x48\x58"+
			"\x9a\x04\x38\x58\xfe")
	r.replace(0x12, 0x5a7b, "nuun portal", "\xf2\x9a\x00",
		"\xf2\xe1\x00\x38\x78\xf3"+nuunPortalOtherObjects+"\xff")
	// and outside D2 present.
	d2PresentPortal := r.appendToBank(0x12, "d2 present portal",
		"\xf2\xdc\x02\x48\x38\xe1\x00\x48\x48\xfe")
	r.replace(0x12, 0x5d42, "d2 present portal pointer",
		"\xdc\x02\x48\x38", "\xf3"+d2PresentPortal+"\xff")

	// bank 15

	// don't equip sword for shooting galleries if player don't have it
	// (doesn't work anyway).
	shootingGalleryEquip := r.appendToBank(0x15, "shooting gallery equip",
		"\x3e\x05\xcd\x48\x17\x3e\x00\x22\xd0\x2b\x3e\x05\x22\xc9")
	r.replace(0x15, 0x50ae, "call shooting gallery equip",
		"\x3e\x05\x22", "\xcd"+shootingGalleryEquip)

	// always make "boomerang" second prize for target carts, checking room
	// flag 6 to track it.
	targetCartsItem := r.appendToBank(0x15, "target carts item",
		"\xcd\x7d\x19\xcb\x77\x3e\x04\xca\xbb\x66\xcd\x3e\x04\xc3\xa5\x66")
	r.replace(0x15, 0x66a2, "call target carts item",
		"\xcd\x3e\x04", "\xc3"+targetCartsItem)
	// set room flag 6 when "boomerang" is given in script.
	targetCartsFlag := r.appendToBank(0x0c, "target carts flag",
		"\xde\x06\x02\xb1\x40\xc1")
	r.replace(0x0c, 0x6e6e, "jump target carts flag",
		"\x88\x6e", targetCartsFlag)

	r.replaceAsm(0x0c, 0x4bd8,
		"db dd,2a,00", "db e0; dw spawnBossItem")

	// bank 16

	r.replaceAsm(0x16, 0x4539,
		"ld b,a; swap a", "call modifyTreasure")

	// bank 21

	// replace ring appraisal text with "you got the {ring}"
	r.replace(0x21, 0x76a0, "obtain ring text replacement",
		"\x04\x2c\x20\x04\x96\x21", "\x02\x06\x0f\xfd\x21\x00")

	// bank 3f

	r.replaceAsm(0x3f, 0x4356,
		"call _interactionGetData", "call checkLoadCustomSprite")
	r.replaceAsm(0x3f, 0x4607,
		"ld hl,4610", "ld hl,seedCapacityTable")
	r.replaceAsm(0x3f, 0x4614,
		"set 6,c; call realignUnappraisedRings", "nop; jp autoAppraiseRing")

	// use different addresses for owl statue text. the text itself is stored
	// in bank $38 instead of $3f, since there's not enough room in $3f.
	owlTextOffsets := r.appendToBank(0x3f, "owl text offsets",
		string(make([]byte, 0x14*2))) // to be set later
	useOwlText := r.appendToBank(0x3f, "use owl text",
		"\xea\xd4\xd0\xfa\xa3\xcb\xfe\x3d\xc0"+ // ret if normal text
			"\x21"+owlTextOffsets+"\xfa\xa2\xcb\xdf\x2a\x66\x6f"+ // set addr
			"\x3e\x38\xea\xd4\xd0\xc9") // set bank
	r.replace(0x3f, 0x4faa, "call use owl text",
		"\xea\xd4\xd0", "\xcd"+useOwlText)

	// this *MUST* be the last thing in the bank, since it's going to grow
	// dynamically later.
	r.appendToBank(0x38, "owl text", "")
}

// makes ages-specific additions to the collection mode table.
func makeAgesCollectModeTable() string {
	b := new(strings.Builder)
	table := makeCollectModeTable()
	b.WriteString(table[:len(table)-1]) // strip final ff

	// add eatern symmetry city brother
	b.Write([]byte{0x03, 0x6f, collectFind2})

	// add ricky and dimitri nuun caves
	b.Write([]byte{0x02, 0xec, collectChest, 0x05, 0xb8, collectChest})

	b.Write([]byte{0xff})
	return b.String()
}
