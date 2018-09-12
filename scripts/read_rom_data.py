#!/usr/bin/env python3

import argparse
import struct
import sys

# because of pyyaml's serialization, we're going to be using lists instead of
# tuples in this script.
import yaml


parser = argparse.ArgumentParser(description="read data from an oos rom.",
        epilog="""
if action is "getroom", two additional hex integer parameters must be
privided for the group ID and room ID of a specific room to get data
from.

if action is "searchchests", an optional hex integer parameters may be
provided to limit the search to a given group ID and music ID.

if action is "searchobjects", two additional hex integer parameters must
be provided for the interaction mode and ID of the objects to search
for. an additional optional hex integer parameters may be provided for
the sub-ID of the objects.

if action is "treasure", two additional hex integer parameters must be
specified for the item ID and sub-ID.
""".strip())
parser.add_argument("romfile", type=str, help="file path of rom to read")
parser.add_argument("action", type=str, help="type of operation to perform")
parser.add_argument("args", type=str, nargs="*", help="action parameters")
args = parser.parse_args()


def fatal(*args):
    print("%s: error:" % __file__, *args, file=sys.stderr)
    exit(2)


def dict_presenter(dumper, data):
    return dumper.represent_dict(data.items())

def hexint_presenter(dumper, data):
    return dumper.represent_int('0x%02x' % data)

yaml.add_representer(dict, dict_presenter)
yaml.add_representer(int, hexint_presenter)


def full_addr(bank_num, offset):
    if bank_num > 2:
        return 0x4000 * (bank_num - 1) + offset
    return offset


MUSIC_PTR_TABLE = 0x04, 0x483c
OBJECT_PTR_TABLE = 0x11, 0x5b3b
CHEST_PTR_TABLE = 0x15, 0x4f6c
TREASURE_PTR_TABLE = 0x15, 0x5129

MUSIC = { # and sound effects
    0x03: "overworld",
    0x0a: "horon village",
    0x0d: "essence room",
    0x0e: "house",
    0x0f: "fairy fountain",
    0x12: "hero's cave",
    0x13: "gnarled root dungeon",
    0x14: "snake's remains",
    0x15: "poison moth's lair",
    0x16: "dancing dragon dungeon",
    0x17: "unicorn's cave",
    0x18: "ancient ruins",
    0x19: "explorer's crypt",
    0x1a: "sword and shield maze",
    0x28: "subrosia",
    0x35: "samasa desert",
    0x36: "cave",
    0x3e: "goron mountain",
    0x4c: "got item",
    0x4d: "puzzle solved (short)",
    0x4e: "damage enemy",
    0x4f: "charge sword",
    0x50: "ping",
    0x51: "shoot rock",
    0x52: "engulf",
    0x53: "jump",
    0x54: "open menu",
    0x55: "close menu",
    0x56: "select option",
    0x57: "restore heart",
    0x58: "deflect",
    0x59: "falling enemy",
    0x5a: "menu says no",
    0x5b: "puzzle solved (long)",
    0x5c: "preparing magic",
    0x5d: "sword beam",
    0x5e: "small key",
    0x4f: "damage link",
    0x60: "low hearts",
    0x70: "onox walk",
    0x80: "minecart",
    0x90: "gale seed",
    0xa0: "dimitri?",
    0xb0: "rumble",
    0xc0: "spell?",
    0xd0: "scent seed impact",
    0xd1: "growl?",
    0xd2: "thunder",
    0xd3: "whirlwind",
}

INTERACTION_MODES = {
    0xf1: "NV interaction",
    0xf2: "DV interaction",
    0xf6: "random entities",
    0xf7: "specific entity",
}

NV_INTERACTIONS = {}

ENTITIES = {
    0x09: ("octorok", {
        0x00: "red 0x00",
        0x01: "red 0x01",
    }),
    0x0a: ("goriya", {
        0x00: "boomerang",
    }),
    0x0e: ("trap", {
        0x00: "spinner",
        0x01: "blade",
    }),
    0x31: ("stalfos", {
        0x00: "blue",
    }),
    0x32: ("keese", {}),
    0x34: ("zol", {
        0x01: "red",
    }),
    0x35: ("floormaster", {}),
    0x38: ("great fairy", {}),
    0x39: ("fire keese", {}),
    0x43: ("gel", {}),
    0x53: ("dragonfly", {}),
    0x59: ("fixed drop", {
        0x00: "fairy",
        0x04: "bombs",
        0x05: "ember seeds",
        0x09: "mystery seeds",
    }),
    0x5a: ("seed tree", {
        0x00: "ember",
        0x01: "mystery",
        0x02: "scent",
        0x03: "pegasus",
        0x04: "gale (sunken city)",
        0x05: "gale (tarm ruins)",
    }),
    0x70: ("goriya bros", {}),
    0x78: ("aquamentus", {}),
}

DV_INTERACTIONS = {
    0x12: ("dungeon", {
        0x00: "entry text",
        0x01: "small key when room cleared",
        0x02: "chest when room cleared",
        0x04: "stairs when room cleared",
    }),
    0x13: ("push block trigger", {}),
    0x1e: ("doors", {
        0x04: "N opens on trigger",
        0x08: "N opens when room cleared",
        0x09: "E opens when room cleared",
        0x0a: "S opens when room cleared",
        0x0b: "W opens when room cleared",
        0x14: "N opens for torches",
        0x15: "W opens for torches",
    }),
    # 0x20 0x01 used for button -> small key chest in d0
    # 0x20 0x02 used for button -> small key chest in d1
    # 0x20 0x03 used for boss room in d1
    # 0x21 and 0x22 are outside the d1 entrance
    # 0x25 0x00 and 0x01 are on the cat-stuck-in-tree screen
    # 0x26 0x00 and 0x01 are also on the cat-stuck-in-tree screen
    # 0x37 0x82 is on the ember tree screen
    0x38: ("d1 old man", {}),
    # 0x44 0x09 is in impa's house
    0x46: ("shopkeeper", {}),
    0x47: ("shop item", {}),
    0x6b: ("placed item", {
        0x0a: "piece of heart",
        # 0x17 is the bridge in the horon subrosia portal cave?
        0x91: "gasha seed",
        0x1f: "gasha seed", # both gasha seeds? maybe set different room flags
        0x20: "seed satchel",
    }),
    0x78: ("toggle tile", {}),
    0x7e: ("miniboss portal", {}),
    0x7f: ("essence", {}),
    0x9d: ("impa", {}),
    0xc6: ("wooden sword", {}),
    0xc7: (0xc7, {
        0x04: "renewable bush",
    }),
    0xdc: ("warp", {
        0x01: "doorway",
        0x02: "chimney",
    }),
    0x31: ("subrosia portal", {}),
    # 0xa5 0x09 used on screen where link falls in the intro
    # 0xdc 0x01 and 0x02 outside hero's cave. entrance ??
    0xe2: ("statue eyes", {}),
}

PARTS = {}

TREASURES = {
    0x00: ("none", {}),
    0x03: ("bombs", {
        0x00: "10 count",
    }),
    0x05: ("sword", {
        0x00: "L-1",
    }),
    0x06: ("boomerang", {
        0x01: "L-2",
    }),
    0x08: ("magnet gloves", {}),
    0x13: ("slingshot", {
        0x00: "L-1",
        0x01: "L-2",
    }),
    0x16: ("power bracelet", {}),
    0x17: ("feather", {
        0x00: "L-1",
        0x01: "L-2",
    }),
    0x28: ("rupees", {
        0x00: "1 count",
        0x01: "5 count",
        0x02: "10 count",
        0x03: "20 count",
        0x04: "30 count",
        0x05: "50 count",
        0x06: "100 count",
    }),
    0x2d: ("ring", {
        0x04: "discovery ring",
        0x05: "moblin ring",
        0x06: "steadfast ring",
        0x07: "rang ring L-1",
        0x08: "blast ring",
        0x09: "quicksand ring",
        0x0a: "quicksand ring",
        0x0b: "armor ring L-2",
        0x0e: "power ring L-1",
        0x10: "subrosian ring",
    }),
    0x2b: ("piece of heart", {}),
    0x30: ("small key", {}),
    0x31: ("boss key", {}),
    0x32: ("compass", {}),
    0x33: ("dungeon map", {}),
    0x34: ("gasha seed", {}),
    0x4f: ("x-shaped jewel", {}),
    0x50: ("red ore", {}),
    0x51: ("blue ore", {}),
    0x54: ("master's plaque", {}),
}


def lookup_entry(table, entry_id, param):
    if entry_id in table:
        entry = table[entry_id]

        if param in entry[1]:
            return [entry_id, param, entry[0], entry[1][param]]

        return [entry_id, param, entry[0]]

    return entry_id, param


def read_byte(buf, bank, addr, increment=0):
    if increment > 0:
        return buf[full_addr(bank, addr)], addr + increment

    return buf[full_addr(bank, addr)]


def read_ptr(buf, bank, addr):
    return struct.unpack_from('<H', buf, offset=full_addr(bank, addr))[0]


def read_music(buf, group, room, name=True):
    bank, addr = MUSIC_PTR_TABLE
    addr = read_ptr(buf, bank, addr + group * 2) + room

    value = read_byte(buf, bank, addr)
    if name:
        if value in MUSIC:
            return MUSIC[value]

    return value


def read_objects(buf, group, room, name=True):
    # read initial pointer
    bank, addr = OBJECT_PTR_TABLE
    addr = read_ptr(buf, bank, addr + group * 2) + room * 2
    addr = read_ptr(buf, bank, addr)

    # read objects (recursively if more pointers are involved)
    objects = []
    while read_byte(buf, bank, addr) != 0xff:
        new_objects, addr = read_interaction(buf, bank, addr, name)
        objects += new_objects

    return objects


def loop_read_interaction(buf, bank, addr, name=True):
    objects = []

    while read_byte(buf, bank, addr) not in (0xfe, 0xff):
        new_objects, addr = read_interaction(buf, bank, addr, name)
        objects += new_objects

    return objects, addr


def read_interaction(buf, bank, addr, name=True):
    objects = []

    # read interaction type
    mode, addr = read_byte(buf, bank, addr, 1)

    if mode == 0xf0:
        print("skipped interaction type", hex(mode), "@", hex(addr - 1),
                file=sys.stderr)
        # TODO
        while read_byte(buf, bank, addr) < 0xf0:
            addr += 1
    elif mode == 0xf1:
        # "no-value interaction"
        nv_interactions = []
        while read_byte(buf, bank, addr) < 0xf0:
            if name:
                kind = list(lookup_entry(NV_INTERACTIONS,
                        read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)))
            else:
                kind = [read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)]
            addr += 2

            objects.append({
                "address": [bank, addr - 2],
                "mode": "NV interaction" if name else mode,
                "variety": kind,
            })
    elif mode == 0xf2:
        # "double-value interaction"
        dv_interactions = []
        while read_byte(buf, bank, addr) < 0xf0:
            if name:
                kind = list(lookup_entry(DV_INTERACTIONS,
                        read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)))
            else:
                kind = [read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)]
            addr += 2

            x, addr = read_byte(buf, bank, addr, 1)
            y, addr = read_byte(buf, bank, addr, 1)

            objects.append({
                "address": [bank, addr - 4],
                "mode": "DV interaction" if name else mode,
                "variety": kind,
                "coords": [x, y],
            })
    elif mode in (0xf3, 0xf4, 0xf5):
        # pointer to other interaction
        ptr = read_ptr(buf, bank, addr)
        addr += 2
        new_objects, _ = loop_read_interaction(buf, bank, ptr, name)
        objects += new_objects
    elif mode == 0xf6:
        # randomly placed entities
        count = read_byte(buf, bank, addr) >> 5
        param = read_byte(buf, bank, addr) & 0x0f
        addr += 1

        if name:
            kind = list(lookup_entry(ENTITIES,
                    read_byte(buf, bank, addr), read_byte(buf, bank, addr+1)))
        else:
            kind = [read_byte(buf, bank, addr), read_byte(buf, bank, addr+1)]
        addr += 2

        objects.append({
            "address": [bank, addr - 3],
            "mode": "random entities" if name else mode,
            "count": count,
            "param": param,
            "variety": kind,
        })
    elif mode == 0xf7:
        # specifically placed entities
        param, addr = read_byte(buf, bank, addr, 1)

        while read_byte(buf, bank, addr) < 0xf0:
            if name:
                kind = list(lookup_entry(ENTITIES,
                        read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)))
            else:
                kind = [read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)]
            addr += 2

            x, addr = read_byte(buf, bank, addr, 1)
            y, addr = read_byte(buf, bank, addr, 1)

            objects.append({
                "address": [bank, addr - 4],
                "mode": "specific entity" if name else mode,
                "param": param,
                "variety": kind,
                "coords": [x, y]
            })
    elif mode == 0xf8:
        while read_byte(buf, bank, addr) < 0xf0:
            if name:
                kind = list(lookup_entry(PARTS,
                        read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)))
            else:
                kind = [read_byte(buf, bank, addr),
                        read_byte(buf, bank, addr+1)]
            addr += 2
            xy, addr = read_byte(buf, bank, addr, 1)

            objects.append({
                "address": [bank, addr - 3],
                "mode": "part" if name else mode,
                "variety": kind,
                "coords": [(xy & 0x0f) * 0x10 + 0x08,
                           ((xy >> 4) & 0x0f)* 0x10 + 0x08]
            })
    elif mode in (0xf9, 0xfa):
        # TODO
        print("skipped interaction type", hex(mode), "@", hex(addr - 1),
                file=sys.stderr)
        while read_byte(buf, bank, addr) < 0xf0:
            addr += 1
    elif mode == 0xfe:
        # end data at pointer
        addr += 1
    elif mode == 0xff:
        # no more interactions to read
        pass
    else:
        print("unknown interaction type", hex(mode), "@", hex(addr - 1),
                file=sys.stderr)

    return objects, addr


def read_chest(buf, group, room):
    # read initial pointer
    bank, addr = CHEST_PTR_TABLE
    addr = read_ptr(buf, bank, addr + group * 2)

    # loop through group chests until marker 0xff is reached.
    # that byte must be used for something else too, but i don't know what.
    while True:
        info, chest_room, treasure_id, treasure_subid = \
                buf[full_addr(bank, addr):full_addr(bank, addr)+4]
        if info == 0xff:
            break

        if chest_room == room:
            return {
                "address": [bank, addr+2],
                "treasure": list(lookup_entry(TREASURES,
                    treasure_id, treasure_subid)),
            }

        addr += 4

    return None


def get_chests(buf, group):
    bank, addr = CHEST_PTR_TABLE
    addr = read_ptr(buf, bank, addr + group * 2)

    # loop through group chests until marker 0xff is reached
    chests = []
    while True:
        info, room, treasure_id, treasure_subid = \
            buf[full_addr(bank, addr):full_addr(bank, addr+4)]
        if info == 0xff:
            break

        chests.append({
            "address": [bank, addr+2],
            "location": [group, room],
            "music": read_music(rom, group, room, name=False),
            "treasure": list(lookup_entry(TREASURES,
                    treasure_id, treasure_subid))
        })

        addr += 4

    return chests


def search_objects(rom, mode, obj_id=None, obj_subid=None):
    # read all interactions in all rooms in all groups, and collate the
    # accumulated objects that match the given ID.
    objects = []
    for group in range(6):
        bank, addr = OBJECT_PTR_TABLE
        addr = read_ptr(rom, bank, addr + group * 2)

        # loop through rooms until the high byte is fxxx, which means that the
        # interaction pointers have ended and the interaction data has started
        for room in range(0x100):
            room_addr = read_ptr(rom, bank, addr + room * 2)

            # read objects (recursively if more pointers are involved)
            room_objects = []
            while read_byte(rom, bank, room_addr) != 0xff:
                new_objects, room_addr = read_interaction(
                        rom, bank, room_addr, name=False)
                room_objects += new_objects

            for obj in room_objects:
                if obj["mode"] == mode:
                    if obj_id is None or  obj["variety"][0] == obj_id:
                        if obj_subid is None or obj["variety"][1] == obj_subid:
                            full_obj = {
                                "location": [group, room],
                                "music": read_music(rom, group, room),
                            }
                            full_obj.update(obj)
                            objects.append(full_obj)

            room += 1

    return objects


def get_treasure(rom, treasure_id, treasure_subid):
    bank, addr = TREASURE_PTR_TABLE
    addr += treasure_id * 4
    if rom[full_addr(bank, addr)] & 0x80:
        addr = read_ptr(rom, bank, addr + 1)
    addr += treasure_subid * 4
    offset = full_addr(bank, addr)
    return [addr] + list(rom[offset:offset+4])


def name_objects(objects):
    for obj in objects:
        obj["mode"] = INTERACTION_MODES[obj["mode"]]
        if obj["mode"] == "NV interaction":
            obj["variety"] = list(lookup_entry(NV_INTERACTIONS,
                    *obj["variety"]))
        elif obj["mode"] in ("random entities", "specific entity"):
            obj["variety"] = list(lookup_entry(ENTITIES, *obj["variety"]))
        elif obj["mode"] in "DV interaction":
            obj["variety"] = list(lookup_entry(DV_INTERACTIONS,
                    *obj["variety"]))


with open(args.romfile, "rb") as f:
    rom = f.read()

if args.action == "getroom":
    if len(args.args) != 2:
        fatal("getroom expects 2 args, got", len(args.args))

    group = int(args.args[0], 16)
    room = int(args.args[1], 16)

    room_data = {
        "group": group,
        "room": room,
        "music": read_music(rom, group, room),
        "objects": read_objects(rom, group, room),
        "chest": read_chest(rom, group, room),
    }

    yaml.dump(room_data, sys.stdout)
elif args.action == "searchchests":
    if len(args.args) == 0: # all groups
        chests = []
        for group in range(8):
            chests += get_chests(rom, group)
    elif len(args.args) == 1: # specific group
        group = int(args.args[0], 16)
        chests = get_chests(rom, group)
    elif len(args.args) == 2: # specific group and music
        group = int(args.args[0], 16)
        music = int(args.args[1], 16)

        chests = get_chests(rom, group)

        # filter by music
        chests = [chest for chest in chests if chest["music"] == music]
    elif len(args.args) > 2:
        fatal("searchchests expects 0-2 args, got", len(args.args))

    # print music by name if known
    for chest in chests:
        if chest["music"] in MUSIC:
            chest["music"] = [chest["music"], MUSIC[chest["music"]]]

    yaml.dump(chests, sys.stdout)
elif args.action == "searchobjects":
    if len(args.args) not in (2, 3):
        fatal("searchobjects expects 2-3 args, got", len(args.args))

    mode = int(args.args[0], 16)
    obj_id = int(args.args[1], 16)
    obj_subid = int(args.args[2], 16) if len(args.args) > 2 else None

    objects = search_objects(rom, mode, obj_id, obj_subid)
    name_objects(objects)

    yaml.dump(objects, sys.stdout)
elif args.action == "treasure":
    if len(args.args) != 2:
        fatal("treasure expects 2 args, got", len(args.args))

    treasure_id = int(args.args[0], 16)
    treasure_subid = int(args.args[1], 16)

    treasure = get_treasure(rom, treasure_id, treasure_subid)

    yaml.dump(treasure, sys.stdout)
elif args.action == "keesanity":
    if len(args.args) != 1:
        fatal("keesanity expects 1 arg, got", len(args.args))

    rand_enemies = search_objects(rom, 0xf6)
    for enemy in rand_enemies:
        addr = full_addr(*enemy["address"])
        rom = rom[:addr] + bytes([0xe0, 0x32, 0x00]) + rom[addr+3:]

    specific_enemies = search_objects(rom, 0xf7)
    for enemy in specific_enemies:
        addr = full_addr(*enemy["address"])
        rom = rom[:addr] + bytes([0x32, 0x00]) + rom[addr+2:]

    with open(args.args[0], 'wb') as f:
        f.write(rom)
else:
    fatal("unknown action:", args.action)