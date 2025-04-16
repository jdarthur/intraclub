import * as React from 'react';
import {useState} from 'react';
import {InputFormItem, NumberInputFormItem, SelectFormItem} from "../../common/FormItem";
import {
    useCreateScoringStructureMutation,
    useGetScoringStructuresQuery,
    useUpdateScoringStructureMutation
} from "../../redux/api.js";
import {CommonFormModal, SubmitResult} from "../../common/CommonFormModal";
import {ScoreCountingType, ScoringStructure, WinCondition} from "../../model/scoring_structure";
import {Form, Space, Steps, Tooltip} from "antd";
import {getCountingTypeLabel, pluralize, scoringStructureDescription} from "./ScoringStructures";
import {InfoCircleOutlined} from "@ant-design/icons";
import {StepFormModal} from "../../common/StepFormModal";
import {StepFormStep} from "../../common/StepForm";

type ScoringStructureFormProps = {
    Update?: boolean // is this updating an existing record or creating a new record
    ScoringStructureId?: string // this will be provided on an update
    InitialState?: ScoringStructure // this will be provided on an update
    ScoreCountingTypes: ScoreCountingType[] // this is used to populate the win condition dropdown
}

export function ScoringStructureForm({
                                         Update,
                                         InitialState,
                                         ScoringStructureId,
                                         ScoreCountingTypes
                                     }: ScoringStructureFormProps) {

    const initialScoreCountingType = getCountingTypeLabel(InitialState.win_condition_counting_type, ScoreCountingTypes)

    const [countingType, setCountingType] = useState<string>(initialScoreCountingType);
    const [showInstantWin, setShowInstantWin] = useState<boolean>(InitialState?.use_instant_win)
    const [winCondition, setWinCondition] = useState<WinCondition>(InitialState.win_condition)
    const [isComposite, setIsComposite] = useState<boolean>(InitialState?.is_composite)


    const [f] = Form.useForm()

    const onChange = (changed: any, values: any) => {
        if (changed?.win_condition_counting_type !== undefined) {
            const t = getCountingTypeLabel(changed.win_condition_counting_type, ScoreCountingTypes)
            setCountingType(t)
        } else if (changed?.use_instant_win !== undefined) {
            setShowInstantWin(changed.use_instant_win)
        } else if (changed?.win_condition !== undefined) {
            setWinCondition(values.win_condition)
        } else if (changed?.is_composite !== undefined) {
            setIsComposite(changed.is_composite)
        }
    }

    const [createScoringStructure] = useCreateScoringStructureMutation()
    const [updateScoringStructure] = useUpdateScoringStructureMutation()

    const scoreCountingTypeOptions = ScoreCountingTypes?.map((s: ScoreCountingType) => ({
        label: `${s.name}s`,
        value: s.type,
    }))

    const onSubmit = async (): Promise<SubmitResult> => {
        const formValues = f.getFieldsValue(true)
        const winCondition: WinCondition = {
            instant_win_threshold: formValues.win_condition.instant_win_threshold,
            must_win_by: formValues.win_condition.must_win_by,
            win_threshold: formValues.win_condition.win_threshold,
        }
        if (showInstantWin !== true) {
            winCondition.instant_win_threshold = 0
        }

        const body: ScoringStructure = {
            name: formValues.name,
            win_condition_counting_type: formValues.win_condition_counting_type,
            win_condition: winCondition,
            secondary_scoring_structures: formValues.secondary_scoring_structures,
        }

        if (isComposite !== true) {
            body.secondary_scoring_structures = []
        }
        let func = () => createScoringStructure(body)
        if (Update) {
            func = () => updateScoringStructure({id: ScoringStructureId, body: body})
        }

        return func();
    }

    const winThresholdSuffix = pluralize(countingType.toLowerCase(), winCondition.win_threshold)
    const mustWinBySuffix = pluralize(countingType.toLowerCase(), winCondition.must_win_by)
    const instantWinThresholdSuffix = pluralize(countingType.toLowerCase(), winCondition.instant_win_threshold)


    const infoStep = <div>
        <InputFormItem name={"name"} label={"Name"}/>
        <SelectFormItem name={"win_condition_counting_type"} label={"Win the match via"}
                        options={scoreCountingTypeOptions}/>
    </div>

    const winConditionStep = <div>
        <NumberInputFormItem name={["win_condition", "win_threshold"]} label={"Win threshold"} min={1}
                             suffix={winThresholdSuffix}/>
        <NumberInputFormItem name={["win_condition", "must_win_by"]} label={"Must win by at least"}
                             min={1} suffix={mustWinBySuffix}/>
        <SelectFormItem name={"use_instant_win"} label={"Has instant win threshold?"}
                        options={[{value: true, label: "Yes"}, {value: false, label: "No"}]}/>
        {showInstantWin ?
            <NumberInputFormItem name={["win_condition", "instant_win_threshold"]}
                                 label={"Instant win threshold"} min={0}
                                 suffix={instantWinThresholdSuffix}/> : null}
    </div>

    const compositeStep = <div>
        <SelectFormItem name={"is_composite"} label={"Is composite?"}
                        options={[{value: true, label: "Yes"}, {value: false, label: "No"}]}/>

        {isComposite ?
            <SelectSecondaryScoringStructure MainWinThreshold={winCondition.win_threshold}
                                             MainWinConditionLabel={countingType}
                                             ScoreCountingTypes={ScoreCountingTypes}
                                             MustWinBy={winCondition.must_win_by}
                                             InstantWinThreshold={winCondition.instant_win_threshold}

            /> : null}
    </div>

    const steps: StepFormStep[] = [
        {title: "Basic info", content: infoStep},
        {title: "Win condition", content: winConditionStep},
        {title: "Composite", content: compositeStep},
    ]

    return <StepFormModal ObjectType={"scoring_structure"} IsUpdate={Update}
                          InitialState={InitialState} form={f} onValuesChange={onChange} steps={steps}
                          onStepFormFinish={onSubmit} children={null} footer={null}/>


}


type SecondaryScoringStructureArgs = {
    MainWinThreshold: number
    MustWinBy: number,
    InstantWinThreshold: number,
    MainWinConditionLabel: string
    ScoreCountingTypes: ScoreCountingType[]
}

function SelectSecondaryScoringStructure({
                                             MainWinThreshold,
                                             MainWinConditionLabel,
                                             ScoreCountingTypes,
                                             MustWinBy,
                                             InstantWinThreshold
                                         }: SecondaryScoringStructureArgs) {
    const {data} = useGetScoringStructuresQuery()
    const options = data?.resource?.map((s: ScoringStructure) => {
        const label = <Space size={"small"}>
            {s.name}
            <Tooltip
                title={scoringStructureDescription(s.win_condition_counting_type, s.win_condition, ScoreCountingTypes)}>
                <InfoCircleOutlined/>
            </Tooltip>
        </Space>

        return {label: label, value: s.id}
    })

    let numberOfMaxMainUnits = (MainWinThreshold * 2) - 1
    if (MustWinBy > 1 && InstantWinThreshold > 0) {
        numberOfMaxMainUnits = (InstantWinThreshold * 2) - 1
    }

    const selects = []
    for (let i = 0; i < numberOfMaxMainUnits; i++) {

        selects.push(<SelectFormItem
            key={`secondary_scoring_structure_${i}`}
            name={["secondary_scoring_structures", i]}
            label={`${MainWinConditionLabel} ${i + 1} win structure`}
            options={options}/>)
    }
    return <div>
        {selects}
    </div>
}




