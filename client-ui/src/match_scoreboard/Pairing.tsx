import * as React from "react"
import {Player, PlayerProps} from "./Player";
import {getLineString, MatchProps, SetScores} from "./SetScores";
import {Card} from "antd";

export type PairingProps = {
    player1: PlayerProps
    player2: PlayerProps
    Home: boolean
    Result?: MatchProps,
    Color: string
    NarrowScreen?: boolean
}

export function Pairing({Result, Home, Color, player1, player2}: PairingProps) {

    const title = <SetScores Us={Home ? Result.Us : Result.Them}
                             Them={Home ? Result.Them : Result.Us}
                             MatchId={"default"}
                             Home={Home}
                             Player1={player1}
                             Player2={player2}
                             PlayoffMode={true}
    />

    const lineString = getLineString(player1, player2)

    return <Card style={{width: "100%", height: "calc(50% - 0.5em)", display: "flex", flexDirection: "column"}}
                 title={title}
                 styles={{
                     header: {
                         padding: 0,
                         fontWeight: "normal",
                     },
                     body: {display: "flex", flex: 1}
                 }}
                 size={"small"}>
        <div style={{
            display: "flex",
            flexDirection: "column",
            width: "100%",
            justifyContent: "space-between",
            padding: "0.25em"
        }}>
            <div style={{display: "flex", paddingBottom: "0.5em", alignItems: "center", flex: 1}}>
                <div style={{width: "50%"}}>
                    <Player player1={true}
                            player_line={player1.line}
                            home={Home}
                            matchup_line={lineString}
                            initialName={player1.name}
                    />
                </div>
                <span style={{
                    margin: "0em 1.5em",
                    height: "100%",
                    border: "1px solid rgba(0, 0, 0, 0.08)"
                }}/>
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
                marginTop: "0.25em",
                background: Color,
                width: "100%",
                height: 5,
                borderRadius: 5,
                border: "1px solid rgba(0, 0, 0, 0.1)",
            }}/>
        </div>

    </Card>

}