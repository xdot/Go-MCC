// Copyright 2017-2019 Andrew Goulas
// https://www.structinf.com
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package gomcc

import (
	"math"
)

// A Location represents the location of an entity in a world.
// Yaw and Pitch are specified in degrees.
type Location struct {
	X, Y, Z, Yaw, Pitch float64
}

const (
	ModelChicken   = "chicken"
	ModelCreeper   = "creeper"
	ModelCrocodile = "croc"
	ModelHumanoid  = "humanoid"
	ModelPig       = "pig"
	ModelPrinter   = "printer"
	ModelSheep     = "sheep"
	ModelSkeleton  = "skeleton"
	ModelSpider    = "spider"
	ModelZombie    = "zombie"
)

type Entity struct {
	server *Server
	client *Client

	id    byte
	name  string
	model string

	DisplayName string
	SkinName    string

	listName  string
	groupName string
	groupRank byte

	level        *Level
	location     Location
	lastLocation Location
}

func NewEntity(name string, server *Server, client *Client) *Entity {
	return &Entity{
		server:      server,
		client:      client,
		id:          0xff,
		name:        name,
		model:       ModelHumanoid,
		DisplayName: name,
		SkinName:    name,
		listName:    name,
	}
}

func (entity *Entity) Server() *Server {
	return entity.server
}

func (entity *Entity) Client() *Client {
	return entity.client
}

func (entity *Entity) ID() byte {
	return entity.id
}

func (entity *Entity) Name() string {
	return entity.name
}

func (entity *Entity) Model() string {
	return entity.model
}

func (entity *Entity) SetModel(model string) {
	if model == entity.model {
		return
	}

	entity.model = model
	if entity.level != nil {
		entity.level.ForEachClient(func(client *Client) {
			client.sendChangeModel(entity)
		})
	}
}

func (entity *Entity) ListName() string {
	return entity.listName
}

func (entity *Entity) Group() string {
	return entity.groupName
}

func (entity *Entity) GroupRank() byte {
	return entity.groupRank
}

func (entity *Entity) SetList(listName string, groupName string, groupRank byte) {
	if listName == entity.listName &&
		groupName == entity.groupName &&
		groupRank == entity.groupRank {
		return
	}

	entity.listName = listName
	entity.groupName = groupName
	entity.groupRank = groupRank
	entity.server.ForEachClient(func(client *Client) {
		client.sendAddPlayerList(entity)
	})
}

func (entity *Entity) Location() Location {
	return entity.location
}

func (entity *Entity) Teleport(location Location) {
	if location == entity.location {
		return
	}

	event := &EventEntityMove{entity, entity.location, location, false}
	entity.server.FireEvent(EventTypeEntityMove, &event)
	if event.Cancel {
		return
	}

	entity.location = location
	if entity.client != nil {
		entity.client.sendTeleport(entity)
	}
}

func (entity *Entity) Level() *Level {
	return entity.level
}

func (entity *Entity) TeleportLevel(level *Level) {
	if entity.level == level {
		return
	}

	lastLevel := entity.level
	if entity.level != nil {
		entity.level = nil
		entity.despawn(lastLevel)
	}

	if level != nil {
		entity.location = level.Spawn
		entity.lastLocation = level.Spawn
		entity.spawn(level)
	}

	entity.level = level

	event := EventEntityLevelChange{entity, lastLevel, level}
	entity.server.FireEvent(EventTypeEntityLevelChange, &event)
}

func (entity *Entity) update() {
	if entity.level == nil {
		return
	}

	positionDirty := false
	if entity.location.X != entity.lastLocation.X ||
		entity.location.Y != entity.lastLocation.Y ||
		entity.location.Z != entity.lastLocation.Z {
		positionDirty = true
	}

	rotationDirty := false
	if entity.location.Yaw != entity.lastLocation.Yaw ||
		entity.location.Pitch != entity.lastLocation.Pitch {
		rotationDirty = true
	}

	teleport := false
	if math.Abs(entity.location.X-entity.lastLocation.X) > 1.0 ||
		math.Abs(entity.location.Y-entity.lastLocation.Y) > 1.0 ||
		math.Abs(entity.location.Z-entity.lastLocation.Z) > 1.0 {
		teleport = true
	}

	var packet interface{}
	if teleport {
		packet = &packetPlayerTeleport{
			packetTypePlayerTeleport,
			entity.id,
			int16(entity.location.X * 32),
			int16(entity.location.Y * 32),
			int16(entity.location.Z * 32),
			byte(entity.location.Yaw * 256 / 360),
			byte(entity.location.Pitch * 256 / 360),
		}
	} else if positionDirty && rotationDirty {
		packet = &packetPositionOrientationUpdate{
			packetTypePositionOrientationUpdate,
			entity.id,
			byte((entity.location.X - entity.lastLocation.X) * 32),
			byte((entity.location.Y - entity.lastLocation.Y) * 32),
			byte((entity.location.Z - entity.lastLocation.Z) * 32),
			byte(entity.location.Yaw * 256 / 360),
			byte(entity.location.Pitch * 256 / 360),
		}
	} else if positionDirty {
		packet = &packetPositionUpdate{
			packetTypePositionUpdate,
			entity.id,
			byte((entity.location.X - entity.lastLocation.X) * 32),
			byte((entity.location.Y - entity.lastLocation.Y) * 32),
			byte((entity.location.Z - entity.lastLocation.Z) * 32),
		}
	} else if rotationDirty {
		packet = &packetOrientationUpdate{
			packetTypeOrientationUpdate,
			entity.id,
			byte(entity.location.Yaw * 256 / 360),
			byte(entity.location.Pitch * 256 / 360),
		}
	} else {
		return
	}

	entity.lastLocation = entity.location
	entity.level.ForEachClient(func(client *Client) {
		if client != entity.client {
			client.sendPacket(packet)
		}
	})
}

func (entity *Entity) Respawn() {
	if entity.level == nil {
		return
	}

	entity.level.ForEachClient(func(client *Client) {
		client.sendDespawn(entity)
	})

	entity.location = entity.level.Spawn
	entity.lastLocation = Location{}

	entity.level.ForEachClient(func(client *Client) {
		client.sendSpawn(entity)
	})
}

func (entity *Entity) spawn(level *Level) {
	level.ForEachClient(func(client *Client) {
		client.sendSpawn(entity)
	})

	if entity.client != nil {
		entity.client.sendLevel(level)
		entity.client.sendSpawn(entity)
		level.ForEachEntity(func(other *Entity) {
			entity.client.sendSpawn(other)
		})
	}
}

func (entity *Entity) despawn(level *Level) {
	level.ForEachClient(func(client *Client) {
		client.sendDespawn(entity)
	})

	if entity.client != nil && entity.client.state == stateGame {
		entity.client.sendDespawn(entity)
		level.ForEachEntity(func(other *Entity) {
			entity.client.sendDespawn(other)
		})
	}
}
