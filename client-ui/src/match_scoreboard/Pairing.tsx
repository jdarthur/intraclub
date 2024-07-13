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

    const m: MatchupProps = {
        // @ts-ignore
        AwayPairing: Home ? {} : pairing,
        AwayTeam: undefined,
        // @ts-ignore
        HomePairing: Home ? pairing : {},
        HomeTeam: {
            name: "",
            color: Color
        },
        Result: Result
    }

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
        <div style={{display: "flex", flexDirection: "row"}}>
            <div style={{
                display: "flex",
                justifyContent: "center",
                marginRight: "0.5em",
                width: NarrowScreen ? "30%" : 30
            }}>
                <div style={{
                    background: Color, width: "min(100%, 50px)", height: "100%",
                    borderRadius: 5,
                    border: "1px solid rgba(0, 0, 0, 0.1)",
                }}/>
            </div>
            <div style={{display: "flex", flexDirection: "column"}}>
                <span style={{marginBottom: "0.5em"}}>
                    <Player player1={true}
                            player_line={player1.line}
                            home={Home}
                            matchup_line={lineString}
                            initialName={player1.name}
                    />
                </span>
                <Player player1={false}
                        player_line={player2.line}
                        home={Home}
                        matchup_line={lineString}
                        initialName={player2.name}
                />
            </div>
        </div>

    </Card>

}