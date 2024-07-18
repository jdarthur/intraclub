import {MatchupProps} from "./Matchup";
import {Card, ColorPicker, Input, Popover} from "antd";
import * as React from "react";
import {calculateScores, CARD_GAP_EM, CARD_WIDTH} from "./Scoreboard";
import {useUpdateTeamInfoMutation} from "../redux/api";
import {useSearchParams} from "react-router-dom";
import {stringEditorDisplayType} from "./Player";
import {HomeOutlined, TruckOutlined} from "@ant-design/icons";
import {WonLostTopLine} from "./WonLostTopLine";

export type Team = {
    name: string
    color: string
}

function TeamColorDisplay({value, setValue, onSave, readOnly}: stringEditorDisplayType) {

    // on change, set the color value on the parent component to the hex of the selected color in the picker
    const onChange = (v: any, hex: string) => {
        setValue(hex)
    }

    // when the modal goes from open -> closed, we will save the data via the API
    const onOpenChange = (open: boolean) => {
        if (!open) {
            console.log("close modal for team color picker")
            onSave()
        }
    }

    return <ColorPicker value={value} onChange={onChange} disabled={readOnly} onOpenChange={onOpenChange}
                        size={"large"}/>
}

function TeamNameDisplay({value, setValue, onSave, readOnly}: stringEditorDisplayType) {

    // on change, set the name value on the parent component to the provided name
    const onChange = (event: any) => {
        setValue(event.target.value)
    }

    // when the modal goes from open -> closed, we will save the data via the API
    const onOpenChange = (open: boolean) => {
        if (!open) {
            console.log("close modal for team name input")
            onSave()
        }
    }

    // name value will show a grey placeholder value when unset
    const nameValue = value ? value : <span style={{color: "#bfbfbf"}}> Name not set </span>

    // this is what's primarily displayed in the modal
    const name = <span style={{minWidth: 100, minHeight: "50px", cursor: readOnly ? "auto" : "pointer"}}>
        {nameValue}
    </span>

    // don't display a popover when we are in read-only mode
    if (readOnly) {
        return name
    }

    // content inside of the popover
    const content = <Input value={value}
                           onChange={onChange}
                           ref={el => {
                               setTimeout(() => el?.focus(), 0); // autofocus the input
                           }}/>

    return <Popover title={"Update team name"} content={content} onOpenChange={onOpenChange} trigger={"click"}>
        {name}
    </Popover>
}

type OneTeamScoreProps = {
    Matchups: MatchupProps[]
    Team: Team,
    Home: boolean
    NarrowScreen?: boolean
}

export function OneTeamScore({Matchups, Team, Home, NarrowScreen}: OneTeamScoreProps) {

    const [team, setTeam] = React.useState<Team>(Team)
    const [updateTeam] = useUpdateTeamInfoMutation()

    // get the `key` value from the query params which determines if we are in read-only mode
    const [searchParams] = useSearchParams()
    const key = searchParams.get('key')

    const setColor = (value: string) => {
        const t: Team = {...team}
        t.color = value
        setTeam(t)
    }

    const setName = (value: string) => {
        const t: Team = {...team}
        t.name = value
        setTeam(t)
    }

    const onSave = () => {
        const body = {
            home: Home,
            name: team.name,
            color: team.color,
            key: key
        }
        console.log(`save team (home=${Home})`, body)
        updateTeam(body)
    }


    return <Card style={{marginBottom: "0.5em", flexGrow: 1}}
                 styles={{body: {padding: "0.25em 1em"}}}>
        <div style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            fontSize: NarrowScreen ? "2.5em" : "3em",
            background: "white",
            padding: NarrowScreen ? "0em" : "0.25em"
        }}>
            <div style={{display: "flex", alignItems: "center"}}>
                <div style={{
                    color: "rgba(0, 0, 0, 0.5)",
                    fontSize: "0.7em",
                    marginRight: "0.5em",
                }}>
                    {Home ? <HomeOutlined style={{marginTop: "0.4em"}}/> :
                        <TruckOutlined style={{marginTop: "0.4em", fontSize: "1.2em"}}/>}
                </div>

                <div style={{marginRight: "0.5em", display: "flex", alignItems: "center"}}>
                    <TeamColorDisplay value={team.color} setValue={setColor} onSave={onSave} readOnly={!key}/>
                </div>

                <div>
                    <TeamNameDisplay value={team.name} setValue={setName} onSave={onSave} readOnly={!key}/>
                </div>

            </div>


            <div style={{fontWeight: "bold", display: "inline-flex", alignItems: "center"}}>
                <span style={{color: "rgba(0, 0, 0, 0.5)", fontSize: "0.6em", marginTop: "0.2em", marginRight: "1em"}}>
                    <WonLostTopLine Matchups={Matchups} Home={Home}/>
                </span>

                {calculateScores(Matchups, Home)}
            </div>
        </div>
    </Card>
}