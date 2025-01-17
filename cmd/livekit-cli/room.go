package main

import (
	"context"
	"fmt"

	"github.com/ggwhite/go-masker"
	livekit "github.com/livekit/protocol/proto"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/urfave/cli/v2"
)

var (
	RoomCommands = []*cli.Command{
		{
			Name:   "create-room",
			Before: createRoomClient,
			Action: createRoom,
			Flags: []cli.Flag{
				urlFlag,
				&cli.StringFlag{
					Name:     "name",
					Usage:    "name of the room",
					Required: true,
				},
				apiKeyFlag,
				secretFlag,
				&cli.StringFlag{
					Name:     "recording-config",
					Usage:    "path to json recording config file",
					Required: false,
				},
			},
		},
		{
			Name:   "list-rooms",
			Before: createRoomClient,
			Action: listRooms,
			Flags: []cli.Flag{
				urlFlag,
				apiKeyFlag,
				secretFlag,
			},
		},
		{
			Name:   "delete-room",
			Before: createRoomClient,
			Action: deleteRoom,
			Flags: []cli.Flag{
				roomFlag,
				urlFlag,
				apiKeyFlag,
				secretFlag,
			},
		},
		{
			Name:   "list-participants",
			Before: createRoomClient,
			Action: listParticipants,
			Flags: []cli.Flag{
				roomFlag,
				urlFlag,
				apiKeyFlag,
				secretFlag,
			},
		},
		{
			Name:   "get-participant",
			Before: createRoomClient,
			Action: getParticipant,
			Flags: []cli.Flag{
				roomFlag,
				identityFlag,
				urlFlag,
				apiKeyFlag,
				secretFlag,
			},
		},
		{
			Name:   "remove-participant",
			Before: createRoomClient,
			Action: removeParticipant,
			Flags: []cli.Flag{
				roomFlag,
				identityFlag,
				urlFlag,
				apiKeyFlag,
				secretFlag,
			},
		},
		{
			Name:   "mute-track",
			Before: createRoomClient,
			Action: muteTrack,
			Flags: []cli.Flag{
				roomFlag,
				identityFlag,
				&cli.StringFlag{
					Name:     "track",
					Usage:    "track sid to mute",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "muted",
					Usage: "set to true to mute, false to unmute",
				},
				urlFlag,
				apiKeyFlag,
				secretFlag,
			},
		},
		{
			Name:   "update-subscriptions",
			Before: createRoomClient,
			Action: updateSubscriptions,
			Flags: []cli.Flag{
				roomFlag,
				identityFlag,
				&cli.StringSliceFlag{
					Name:     "track",
					Usage:    "track sid to subscribe/unsubscribe",
					Required: true,
				},
				&cli.BoolFlag{
					Name:  "subscribe",
					Usage: "set to true to subscribe, otherwise it'll unsubscribe",
				},
				urlFlag,
				apiKeyFlag,
				secretFlag,
			},
		},
	}

	roomClient *lksdk.RoomServiceClient
)

func createRoomClient(c *cli.Context) error {
	url := c.String("url")
	apiKey := c.String("api-key")
	apiSecret := c.String("api-secret")

	if c.Bool("verbose") {
		fmt.Printf("creating client to %s, with api-key: %s, secret: %s\n",
			url,
			masker.ID(apiKey),
			masker.ID(apiSecret))
	}

	roomClient = lksdk.NewRoomServiceClient(url, apiKey, apiSecret)
	return nil
}

func createRoom(c *cli.Context) error {
	room, err := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name: c.String("name"),
	})
	if err != nil {
		return err
	}

	PrintJSON(room)
	return nil
}

func listRooms(c *cli.Context) error {
	res, err := roomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	if err != nil {
		return err
	}
	if len(res.Rooms) == 0 {
		fmt.Println("there are no active rooms")
	}
	for _, rm := range res.Rooms {
		fmt.Printf("%s\t%s\n", rm.Sid, rm.Name)
	}
	return nil
}

func deleteRoom(c *cli.Context) error {
	roomId := c.String("room")
	_, err := roomClient.DeleteRoom(context.Background(), &livekit.DeleteRoomRequest{
		Room: roomId,
	})
	if err != nil {
		return err
	}

	fmt.Println("deleted room", roomId)
	return nil
}

func listParticipants(c *cli.Context) error {
	roomName := c.String("room")
	res, err := roomClient.ListParticipants(context.Background(), &livekit.ListParticipantsRequest{
		Room: roomName,
	})
	if err != nil {
		return err
	}

	for _, p := range res.Participants {
		fmt.Printf("%s (%s)\t tracks: %d\n", p.Identity, p.State.String(), len(p.Tracks))
	}
	return nil
}

func getParticipant(c *cli.Context) error {
	roomName := c.String("room")
	identity := c.String("identity")
	res, err := roomClient.GetParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	if err != nil {
		return err
	}

	PrintJSON(res)

	return nil
}

func removeParticipant(c *cli.Context) error {
	roomName := c.String("room")
	identity := c.String("identity")
	_, err := roomClient.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	if err != nil {
		return err
	}

	fmt.Println("successfully removed participant", identity)

	return nil
}

func muteTrack(c *cli.Context) error {
	roomName := c.String("room")
	identity := c.String("identity")
	trackSid := c.String("track")
	_, err := roomClient.MutePublishedTrack(context.Background(), &livekit.MuteRoomTrackRequest{
		Room:     roomName,
		Identity: identity,
		TrackSid: trackSid,
		Muted:    c.Bool("muted"),
	})
	if err != nil {
		return err
	}

	verb := "muted"
	if !c.Bool("muted") {
		verb = "unmuted"
	}
	fmt.Println(verb, "track: ", trackSid)
	return nil
}

func updateSubscriptions(c *cli.Context) error {
	roomName := c.String("room")
	identity := c.String("identity")
	trackSids := c.StringSlice("track")
	_, err := roomClient.UpdateSubscriptions(context.Background(), &livekit.UpdateSubscriptionsRequest{
		Room:      roomName,
		Identity:  identity,
		TrackSids: trackSids,
		Subscribe: c.Bool("subscribe"),
	})
	if err != nil {
		return err
	}

	verb := "subscribed to"
	if !c.Bool("subscribe") {
		verb = "unsubscribed from"
	}
	fmt.Println(verb, "tracks: ", trackSids)
	return nil
}
