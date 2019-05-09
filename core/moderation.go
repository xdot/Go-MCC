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

package core

import (
	"net"
	"strings"

	"Go-MCC/gomcc"
)

var commandBan = gomcc.Command{
	Name:        "ban",
	Description: "Ban a player from the server.",
	Permission:  "core.ban",
	Handler:     handleBan,
}

func handleBan(sender gomcc.CommandSender, command *gomcc.Command, message string) {
	if len(message) == 0 {
		sender.SendMessage("Usage: " + command.Name + " <player> <reason>")
		return
	}

	reason := "You have been banned"
	args := strings.SplitN(message, " ", 2)
	if len(args) > 1 {
		reason = args[1]
	}

	if !gomcc.IsValidName(args[0]) {
		sender.SendMessage(args[1] + " is not a valid name")
		return
	}

	if CoreBans.Name.Ban(args[0], reason, sender.Name()) {
		sender.SendMessage("Player " + args[0] + " banned")
		player := sender.Server().FindPlayer(args[0])
		if player != nil {
			player.Kick(reason)
		}
	} else {
		sender.SendMessage("Player " + args[0] + " is already banned")
	}
}

var commandBanIp = gomcc.Command{
	Name:        "banip",
	Description: "Ban an IP address from the server.",
	Permission:  "core.banip",
	Handler:     handleBanIp,
}

func handleBanIp(sender gomcc.CommandSender, command *gomcc.Command, message string) {
	if len(message) == 0 {
		sender.SendMessage("Usage: " + command.Name + " <ip> <reason>")
		return
	}

	reason := "You have been banned"
	args := strings.SplitN(message, " ", 2)
	if len(args) > 1 {
		reason = args[1]
	}

	if net.ParseIP(args[0]) == nil {
		sender.SendMessage(args[0] + " is not a valid IP address")
		return
	}

	if CoreBans.IP.Ban(args[0], reason, sender.Name()) {
		sender.SendMessage("IP " + args[0] + " banned")
		sender.Server().ForEachPlayer(func(player *gomcc.Player) {
			if player.RemoteAddr() == args[0] {
				player.Kick(reason)
			}
		})
	} else {
		sender.SendMessage("IP " + args[0] + " is already banned")
	}
}

var commandKick = gomcc.Command{
	Name:        "kick",
	Description: "Kick a player from the server.",
	Permission:  "core.kick",
	Handler:     handleKick,
}

func handleKick(sender gomcc.CommandSender, command *gomcc.Command, message string) {
	if len(message) == 0 {
		sender.SendMessage("Usage: " + command.Name + " <player> <reason>")
		return
	}

	args := strings.SplitN(message, " ", 2)
	player := sender.Server().FindPlayer(args[0])
	if player == nil {
		sender.SendMessage("Player " + args[0] + " not found")
		return
	}

	reason := "Kicked by " + sender.Name()
	if len(args) > 1 {
		reason = args[1]
	}

	player.Kick(reason)
}

var commandRank = gomcc.Command{
	Name:        "rank",
	Description: "Set the rank of a player.",
	Permission:  "core.rank",
	Handler:     handleRank,
}

func handleRank(sender gomcc.CommandSender, command *gomcc.Command, message string) {
	args := strings.Fields(message)
	switch len(args) {
	case 1:
		if player := CorePlayers.OfflinePlayer(args[0]); player != nil {
			sender.SendMessage("The rank of " + args[0] + " is " + player.Rank)
		} else {
			sender.SendMessage("Player " + args[0] + " not found")
		}

	case 2:
		if player := CorePlayers.OfflinePlayer(args[0]); player != nil {
			if CoreRanks.Rank(args[1]) == nil {
				sender.SendMessage("Rank " + args[1] + " not found")
				return
			}

			player.Rank = args[1]
			sender.SendMessage("Rank of " + args[0] + " set to " + args[1])

			if player := CorePlayers.Player(args[0]); player != nil {
				player.UpdatePermissions()
			}
		} else {
			sender.SendMessage("Player " + args[0] + " not found")
		}

	default:
		sender.SendMessage("Usage: " + command.Name + " <player> <rank>")
	}
}

var commandUnban = gomcc.Command{
	Name:        "unban",
	Description: "Remove the ban for a player.",
	Permission:  "core.unban",
	Handler:     handleUnban,
}

func handleUnban(sender gomcc.CommandSender, command *gomcc.Command, message string) {
	args := strings.Fields(message)
	if len(args) != 1 {
		sender.SendMessage("Usage: " + command.Name + " <player>")
		return
	}

	if CoreBans.Name.Unban(args[0]) {
		sender.SendMessage("Player " + args[0] + " unbanned")
	} else {
		sender.SendMessage("Player " + args[0] + " is not banned")
	}
}

var commandUnbanIp = gomcc.Command{
	Name:        "unbanip",
	Description: "Remove the ban for an IP address.",
	Permission:  "core.unbanip",
	Handler:     handleUnbanIp,
}

func handleUnbanIp(sender gomcc.CommandSender, command *gomcc.Command, message string) {
	args := strings.Fields(message)
	if len(args) != 1 {
		sender.SendMessage("Usage: " + command.Name + " <ip>")
		return
	}

	if CoreBans.IP.Unban(args[0]) {
		sender.SendMessage("IP " + args[0] + " unbanned")
	} else {
		sender.SendMessage("IP " + args[0] + " is not banned")
	}
}
