import * as React from "react"
import {CheckCircleOutlined, CloseCircleOutlined} from "@ant-design/icons";
import {PlayerProps} from "./Player";
import {OneSetScore} from "./OneSetScore";
import {useSearchParams} from "react-router-dom";
import {useUpdateMatchScoresForLineMutation} from "../redux/api.js";
import {useEffect} from "react";
import {IntraclubTokenKey, setCredentials} from "../redux/auth";

export type SetScoreProps = {
    set1_games: number,
    set2_games: number,
    set3_games?: number,
}

export type MatchProps = {
    Us: SetScoreProps
    Them: SetScoreProps
    PlayoffMode?: boolean
    MatchId: string
    Home?: boolean
    Player1: PlayerProps
    Player2: PlayerProps
}

export function getLineString(player1: PlayerProps, player2: PlayerProps): string {
    return `${player1.line}-${player2.line}`
}

function isMatchWon(us: SetScoreProps, them: SetScoreProps): boolean {
    if (us.set3_games > 0) {
        return true
    }

    const set1Won = (us.set1_games == 6 && them.set1_games < 5) || us.set1_games == 7
    const set2Won = (us.set2_games == 6 && them.set2_games < 5) || us.set2_games == 7

    return set1Won && set2Won;
}

function scoreStyle(last: boolean, won: boolean, lost: boolean) {
    const x = {
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        flexGrow: 1,
    }
    if (!last) {
        x["borderRight"] = "1px solid rgba(0, 0, 0, 0.1)"
    } else {
        x["fontWeight"] = "bold"
        x["fontSize"] = "1.1em"
        x["borderRadius"] = "0px 7px 0px 0px"
        x["padding"] = "0.5em"

        if (won) {
            x["background"] = "#d9f7be"
        }
        if (lost) {
            x["background"] = "#ffccc7"
        }

    }
    return x
}

export function calcTotal(match: MatchProps, home: boolean): number {

    const us = home ? match.Us : match.Them
    const them = home ? match.Them : match.Us

    let total = us.set1_games + us.set2_games + us.set3_games

    const won = isMatchWon(us, them)

    if (won) {
        total += 5
    }
    return total
}

type SetValues = {
    Set1Value: number
    Set2Value: number
    Set3Value: number,
    PlayoffMode?: boolean
}

export function SetScores({Us, Them, MatchId, Home, PlayoffMode, Player1, Player2}: MatchProps) {
    const [setValues, updateSetValues] = React.useState<SetValues>({
        PlayoffMode: false,
        Set1Value: Us.set1_games,
        Set2Value: Us.set2_games,
        Set3Value: Us.set3_games,
    })

    useEffect(() => {
        const v: SetValues = {
            PlayoffMode: false,
            Set1Value: Us.set1_games,
            Set2Value: Us.set2_games,
            Set3Value: Us.set3_games,
        }
        updateSetValues(v)

    }, [Us])

    const [updateSetScores] = useUpdateMatchScoresForLineMutation()

    // get the `key` value from the query params which determines if we are in read-only mode
    const [searchParams] = useSearchParams()
    const key = searchParams.get('key')

    const won = isMatchWon(Us, Them)
    const lost = isMatchWon(Them, Us)

    const match: MatchProps = {
        Home: Home,
        MatchId: MatchId,
        Player1: Player1,
        Player2: Player2,
        PlayoffMode: PlayoffMode,
        Them: Us,
        Us: Them,
    }

    const total = calcTotal(match, false)

    let wonLostDisplay = null
    if (won) {
        wonLostDisplay = <CheckCircleOutlined/>
    } else if (lost) {
        wonLostDisplay = <CloseCircleOutlined/>
    }

    const totalDisplay = <div style={{display: "flex", alignItems: "center"}}>
        {total}
        <span style={{width: "0.5em"}}/>
        {wonLostDisplay}
    </div>

    const onSave = () => {
        let body = {
            scores: {
                set1_games: setValues.Set1Value,
                set2_games: setValues.Set2Value,
                set3_games: setValues.Set3Value,
            },
            line: getLineString(Player1, Player2),
            home: Home,
            key: key
        }

        console.log(`Save values for matchID ${MatchId}, ${getLineString(Player1, Player2)} (home=${Home}): `, body)
        updateSetScores(body)
    }

    const updateSetValue = (index: number, value: number) => {
        const v: SetValues = {...setValues}
        if (index == 0) {
            v.Set1Value = value
        } else if (index == 1) {
            v.Set2Value = value
        } else if (index == 2) {
            v.Set3Value = value
        } else {
            console.log("invalid set index: ", index)
        }

        updateSetValues(v)
    }

    return <div
        style={{
            display: "flex",
            flexDirection: "row",
            justifyContent: "space-evenly",
            alignItems: "center",
            fontSize: "1.6em"
        }}>
        <div style={scoreStyle(false, false, false)}>
            Scores
        </div>
        <div style={scoreStyle(false, false, false)}>
            <OneSetScore
                value={setValues.Set1Value}
                max={Them.set1_games == 5 ? 7 : 6}
                setValue={(v) => updateSetValue(0, v)}
                onSave={onSave}
                readOnly={!key}
            />
        </div>
        <div style={scoreStyle(false, false, false)}>
            <OneSetScore
                value={setValues.Set2Value}
                max={Them.set2_games == 5 ? 7 : 6}
                setValue={(v) => updateSetValue(1, v)}
                onSave={onSave}
                readOnly={!key}/>
        </div>

        <div style={scoreStyle(false, false, false)}>
            <OneSetScore
                value={setValues.Set3Value}
                max={1}
                setValue={(v) => updateSetValue(2, v)}
                onSave={onSave}
                readOnly={!key}/>
        </div>
        <div style={scoreStyle(false, false, false)}>{won ? "+5" : "-"}</div>
        <div style={scoreStyle(true, won, lost)}>{totalDisplay}</div>
    </div>
}


