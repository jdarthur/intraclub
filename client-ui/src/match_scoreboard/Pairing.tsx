import * as React from "react"
import {Player, PlayerProps} from "./Player";
import {getLineString, MatchProps, SetScores} from "./SetScores";
import {Card} from "antd";
import {MatchupProps} from "./Matchup";

export type PairingProps = {
    player1: PlayerProps
    player2: PlayerProps
    Home: boolean
    Result?: MatchProps,
    Color: string
    NarrowScreen?: boolean
}

export function Pairing({Result, Home, Color, NarrowScreen, player1, player2}: PairingProps) {

    const pairing: PairingProps = {Color: Color, Home: Home, player1: player1, player2: player2}

    const title = <SetScores Us={Home ? Result.Us : Result.Them}
                             Them={Home ? Result.Them : Result.Us}
                             MatchId={"default"}
                             Home={Home}
                             Player1={player1}
                             Player2={player2}
                             PlayoffMode={true}
    />

    const lineString = getLineString(player1, player2)

    return <Card style={{width: "100%"}} title={title}
                 styles={{header: {padding: 0, fontWeight: "normal"}}}
                 size={"small"}>
        <div style={{
            display: "flex",
            flexDirection: "column",
            width: "100%",
            justifyContent: "space-between",
        }}>
            <div style={{display: "flex", paddingBottom: "0.5em"}}>
                <div style={{width: "50%"}}>
                    <Player player1={true}
                            player_line={player1.line}
                            home={Home}
                            matchup_line={lineString}
                            initialName={player1.name}
                    />
                </div>
                <div style={{width: "50%"}}>
                    <Player player1={false}
                            player_line={player2.line}
                            home={Home}
                            matchup_line={lineString}
                            initialName={player2.name}
                    />
                </div>
            </div>
            <div style={{
                display: "flex",
                justifyContent: "center",
                marginRight: "1em",
                width: "100%",
            }}>
                <div style={{
                    background: Color, width: "100%",
                    height: 5,
                    borderRadius: 5,
                    border: "1px solid rgba(0, 0, 0, 0.1)",
                }}/>
            </div>
        </div>

    </Card>

}