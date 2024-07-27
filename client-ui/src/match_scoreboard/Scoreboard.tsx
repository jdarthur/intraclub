import * as React from "react";
import {MatchupProps} from "./Matchup";
import {Skeleton} from "antd";
import {calcTotal, MatchProps} from "./SetScores";
import {PairingProps} from "./Pairing";
import {useLazyGetMatchScoresQuery} from "../redux/api.js";
import {PlayerProps} from "./Player";
import {OneTeamScore} from "./OneTeamScore";
import {MatchupGroup} from "./MatchupGroup";

export const CARD_WIDTH = 300
export const CARD_GAP_EM = 2
const WEBSOCKET_PING_INTERVAL_MS = 60000 // * 15

export function calculateScores(matchups: MatchupProps[], home: boolean): number {
    let total = 0
    for (let i = 0; i < matchups.length; i++) {
        const m = matchups[i]
        const m2: MatchProps = {
            Player1: undefined, Player2: undefined,
            MatchId: "default",
            Us: m.Result.Us,
            Them: m.Result.Them
        }
        total += calcTotal(m2, home)
    }
    return total
}


function emptyPairing(color: string, home: boolean): PairingProps {
    return {
        Color: color,
        Home: home,
        player1: emptyPlayer(),
        player2: emptyPlayer()
    }
}

function emptyPlayer(): PlayerProps {
    return {
        line: 0,
        name: ""
    }
}

function getMatchup(object: any, index: number, home: boolean): MatchupProps {
    const m: MatchupProps = {
        AwayPairing: emptyPairing(object?.away_team.color, false),
        AwayTeam: object?.away_team,
        HomePairing: emptyPairing(object?.home_team.color, true),
        HomeTeam: object?.home_team,
        Result: getResult(object, index, home)
    }

    m.HomePairing.player1 = getPlayer(object, true, true, index)
    m.HomePairing.player2 = getPlayer(object, true, false, index)

    m.AwayPairing.player1 = getPlayer(object, false, true, index)
    m.AwayPairing.player2 = getPlayer(object, false, false, index)

    return m
}

function getAllMatchups(object: any): MatchupProps[] {
    if (!object) {
        return []
    }

    const output: MatchupProps[] = []
    for (let i = 0; i < 6; i++) {
        const m = getMatchup(object, i, true)
        output.push(m)
    }

    return output
}

function getPlayer(object: any, home: boolean, player1: boolean, index: number): PlayerProps {
    const player = emptyPlayer()
    const v = getLineValue(object, home, index)
    if (player1) {
        player.name = v?.pairing.player1.name
        player.line = v?.pairing.player1.line
    } else {
        player.name = v?.pairing.player2.name
        player.line = v?.pairing.player2.line
    }
    return player
}

type LineValue = {
    pairing: PairingProps
    set_scores: {
        set1_games: number,
        set2_games: number,
        set3_games: number,
    }
}

function getLineValue(object: any, home: boolean, index: number): LineValue {
    if (!object) {
        const p: PairingProps = {
            Color: "", Home: false, player1: {
                line: 0,
                name: ""
            }, player2: {
                line: 0,
                name: ""
            }
        }
        return {pairing: p, set_scores: {set1_games: 0, set2_games: 0, set3_games: 0}}
    }

    let root = object?.home_scores
    if (!home) {
        root = object?.away_scores
    }

    if (index == 0) {
        return root?.one_one
    } else if (index == 1) {
        return root?.one_two
    } else if (index == 2) {
        return root?.one_three
    } else if (index == 3) {
        return root?.two_two
    } else if (index == 4) {
        return root?.two_three
    } else if (index == 5) {
        return root?.three_three
    } else {
        console.error("invalid line index: ", index)
        return {pairing: undefined, set_scores: {set1_games: 0, set2_games: 0, set3_games: 0}}
    }

}

function getResult(object: any, index: number, home: boolean): MatchProps {
    return {
        Player1: getPlayer(object, home, true, index),
        Player2: getPlayer(object, home, false, index),
        Home: home,
        MatchId: "",
        PlayoffMode: true,
        Them: getLineValue(object, !home, index).set_scores,
        Us: getLineValue(object, home, index).set_scores
    }
}

export function Scoreboard() {

    const [width, setWidth] = React.useState(window.innerWidth);

    let socket: WebSocket = null

    const breakpoint = CARD_WIDTH * 3;
    React.useEffect(() => {
        const handleResizeWindow = () => {
            setWidth(window.innerWidth);
        }

        // subscribe to window resize event "onComponentDidMount"
        window.addEventListener("resize", handleResizeWindow);

        websocketFunc()

        setInterval(() => {
            console.log("send ping")
            socket.send("p")
        }, WEBSOCKET_PING_INTERVAL_MS)

        return () => {
            // unsubscribe "onComponentDestroy"
            window.removeEventListener("resize", handleResizeWindow);
        };
    }, []);

    const websocketFunc = () => {
        socket = new WebSocket(`ws://${window.location.host}/api/ws`);

        // Connection opened
        socket.addEventListener("open", (event) => {
            trigger()
        });

        // Listen for messages
        socket.addEventListener("message", (event) => {
            trigger()
        });

        socket.onclose = () => {
            console.log("socket closed")
        }

    }

    const narrowScreen = width < breakpoint

    const [trigger, result] = useLazyGetMatchScoresQuery()
    const data = result?.data
    if (result.isUninitialized || result.isLoading) {
        return <Skeleton/>
    }

    const HomeTeam = data?.home_team
    const AwayTeam = data?.away_team

    const Matchups = getAllMatchups(data)

    const mainStyle = {
        height: "100%", overflowY: "auto",
    }

    const row1 = [
        getMatchup(data, 0, true),
        getMatchup(data, 1, true),
        getMatchup(data, 2, true),
    ]

    const row2 = [
        getMatchup(data, 3, true),
        getMatchup(data, 4, true),
        getMatchup(data, 5, true),
    ]


    // @ts-ignore
    return <div style={mainStyle}>
        <div style={{
            display: "block",
            flexDirection: "column",
            flexWrap: "wrap",
            overflowY: narrowScreen ? "auto" : "clip",
            justifyContent: "space-around",
            width: "100%",
            height: narrowScreen ? "" : "100%"
        }}>
            <div style={{
                padding: "1em",
                paddingBottom: narrowScreen ? "0em" : "0.5em",
                display: "flex",
                width: "calc(100% - 2em)",
                flexDirection: narrowScreen ? "column" : "row",
                justifyContent: narrowScreen ? "flex-start" : "space-between",
            }}>
                <OneTeamScore Matchups={Matchups}
                              Team={HomeTeam}
                              Home={true}
                              NarrowScreen={narrowScreen}/>
                <span style={{width: narrowScreen ? "0.25em" : "1em"}}/>
                <OneTeamScore Matchups={Matchups}
                              Team={AwayTeam}
                              NarrowScreen={narrowScreen}
                              Home={false}
                />
            </div>
            <div style={{height: "90%", width: "100%"}}>
                <MatchupGroup Matchups={row1} NarrowScreen={narrowScreen} ScreenWidth={width}/>
                <div style={{height: "2em"}}/>
                <MatchupGroup Matchups={row2} NarrowScreen={narrowScreen} ScreenWidth={width}/>
            </div>

        </div>
    </div>
}