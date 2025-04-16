import * as React from 'react';
import {
    useDeleteScoringStructureMutation,
    useGetScoreCountingTypesQuery,
    useGetScoringStructuresQuery
} from "../../redux/api.js";
import {ScoreCountingType, ScoringStructure, WinCondition} from "../../model/scoring_structure";
import {Ellipsis} from "../../common/Ellipsis";
import {Card, Empty, Space} from "antd";
import {DeleteConfirm} from "../../common/DeleteConfirm";
import {LabeledValue} from "../../common/LabeledValue";
import {ScoringStructureForm} from "./ScoringStructureForm";

function scoringStructureDefaults(): ScoringStructure {
    return {
        id: "",
        name: "",
        owner: "",
        secondary_scoring_structures: [],
        win_condition: {
            win_threshold: 1,
            must_win_by: 1,
            instant_win_threshold: 0
        },
        use_instant_win: false,
        is_composite: false,
        win_condition_counting_type: 0
    }
}

export function ScoringStructures() {
    const {data: scoreCountingTypes} = useGetScoreCountingTypesQuery()

    const {data} = useGetScoringStructuresQuery()

    const scoringStructures = data?.resource?.map((s: ScoringStructure) => (
        <OneScoringStructure key={s.id} id={s.id} owner={s.owner} name={s.name}
                             win_condition_counting_type={s.win_condition_counting_type}
                             win_condition={s.win_condition}
                             secondary_scoring_structures={s.secondary_scoring_structures}
                             ScoreCountingTypes={scoreCountingTypes?.resource}/>
    ))

    console.log("mount scoring structures")

    return <div style={{display: "flex", flexWrap: "wrap", gap: "1em"}}>
        {scoringStructures?.length ? scoringStructures : <Empty/>}
        <div style={{height: "1em"}}/>
        {scoreCountingTypes ?
            <ScoringStructureForm ScoreCountingTypes={scoreCountingTypes?.resource}
                                  InitialState={scoringStructureDefaults()}/> : null}
    </div>
}

type ScoringStructureArgs = ScoringStructure & {
    ScoreCountingTypes: ScoreCountingType[],
}

export function OneScoringStructure({
                                        id,
                                        owner,
                                        name,
                                        win_condition_counting_type,
                                        win_condition,
                                        secondary_scoring_structures,
                                        ScoreCountingTypes
                                    }: ScoringStructureArgs) {

    const [deleteScoringStructure] = useDeleteScoringStructureMutation()


    const deleteSelf = (): string => {
        return deleteScoringStructure(id).then((res: { error: any; data: any; }) => {
            return res.error ? res.error : ""
        });
    }

    const initialState: ScoringStructure = {
        id, owner, name, win_condition_counting_type, win_condition, secondary_scoring_structures
    }
    initialState.is_composite = secondary_scoring_structures ? secondary_scoring_structures?.length > 0 : false
    initialState.use_instant_win = win_condition.instant_win_threshold > 0

    const title = <Ellipsis fullValue={name} maxLength={25}/>

    const editForm = <ScoringStructureForm ScoringStructureId={id}
                                           InitialState={initialState} Update
                                           ScoreCountingTypes={ScoreCountingTypes}/>

    const extra = <Space>
        {editForm}
        <DeleteConfirm deleteFunction={deleteSelf} objectType={"facility"}/>
    </Space>

    return <Card title={title} style={{width: 250}} size={"small"}
                 extra={extra}>
        <LabeledValue label={"Composite"} value={secondary_scoring_structures?.length > 0 ? "yes" : "no"}/>
        <LabeledValue label={"Description"}
                      value={scoringStructureDescription(win_condition_counting_type, win_condition, ScoreCountingTypes)}
                      vertical/>

    </Card>
}


export function getCountingTypeLabel(numeric: number, all: ScoreCountingType[]): string {
    const label = all?.find(e => e.type === numeric)
    if (!label) {
        console.log("did not find counting type: ", numeric, all)
        return ""
    }
    return label.name
}

export function scoringStructureDescription(win_condition_counting_type: number, win_condition: WinCondition, ScoreCountingTypes: ScoreCountingType[]): any {

    const typeLabel = getCountingTypeLabel(win_condition_counting_type, ScoreCountingTypes)
    let output = `First to ${win_condition?.win_threshold} ${pluralize(typeLabel.toLowerCase(), win_condition.win_threshold)}`
    if (win_condition?.must_win_by > 1) {
        output += `, win-by-${win_condition.must_win_by}`
    }
    if (win_condition.instant_win_threshold) {
        output += ` (or first to ${win_condition.instant_win_threshold} ${pluralize(typeLabel.toLowerCase(), win_condition.instant_win_threshold)})`
    }

    return <i>{output}</i>

}

export function pluralize(str: string, number: number): string {
    if (str === "") {
        return ""
    }

    if (number === 1) {
        return str
    }
    return `${str}s`

}

