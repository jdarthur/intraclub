import * as React from 'react';
import {useState} from 'react';
import {DoubleSelectFormItem, InputFormItem, NumberInputFormItem, SelectFormItem,} from "../../common/FormItem";
import {useCreateFormatMutation, useGetRatingsQuery, useUpdateFormatMutation} from "../../redux/api.js";
import {SubmitResult} from "../../common/CommonFormModal";
import {Form, Spin} from "antd";
import {StepFormStep} from "../../common/StepForm";
import {StepFormModal} from "../../common/StepFormModal";
import {RatingTooltip} from "./Formats";

type FormatFormProps = {
    Update?: boolean // is this updating an existing record or creating a new record?
    FormatId?: string // this will be provided on an update
    InitialState?: Format // this will be provided on an update
}

export function FormatForm({Update, InitialState, FormatId}: FormatFormProps) {
    const [createFormat] = useCreateFormatMutation()
    const [updateFormat] = useUpdateFormatMutation()

    const [lineCount, setLineCount] = useState<number>(0)
    const [possibleRatings, setPossibleRatings] = useState<string[]>([])

    const [f] = Form.useForm()

    const onSubmit = async (): Promise<SubmitResult> => {
        const values = f.getFieldsValue(true)
        const body: Format = {
            name: values.name,
            possible_ratings: values.possible_ratings,
            lines: values.lines,
        }
        let func = () => createFormat(body)
        if (Update) {
            func = () => updateFormat({id: FormatId, body: body})
        }
        return func();
    }

    const onChange = (changed: any, values: any) => {
        console.log("changed", changed)
        if (changed?.line_count !== undefined) {
            setLineCount(changed?.line_count)
        }
        if (changed?.possible_ratings !== undefined) {
            setPossibleRatings(changed?.possible_ratings)
        }
    }

    const infoStep = <div>
        <InputFormItem name={"name"} label={"Name"}/>
        <RatingSelect/>
    </div>

    // the max number of lines is e.g., 3+2+1 or 4+3+2+1
    const possibleRatingCount = possibleRatings?.length
    const maxNumberOfLines = possibleRatingCount * (possibleRatingCount + 1) * 2
    const selects = []
    for (let i = 0; i < lineCount; i++) {
        selects.push(<LineSelect key={i} PossibleRatings={possibleRatings} Index={i}/>)
    }
    const lines = <div>
        <NumberInputFormItem name={"line_count"} label={"Number of lines"} max={maxNumberOfLines}/>
        {selects}
    </div>

    const steps: StepFormStep[] = [
        {title: "Basic info", content: infoStep},
        {title: "Lines", content: lines, disabled: possibleRatingCount === 0},
    ]

    return <StepFormModal ObjectType={"format"} IsUpdate={Update}
                          InitialState={InitialState} form={f} onValuesChange={onChange} steps={steps}
                          onStepFormFinish={onSubmit} children={null} footer={null}/>
}

export function RatingSelect() {
    const {data, isFetching} = useGetRatingsQuery()
    if (isFetching) {
        return <Spin/>
    }

    const options = data?.resource?.map((r: Rating) => {
        const label = <RatingTooltip RatingId={r.id}/>
        return {label: label, value: r.id}
    })

    return <SelectFormItem name={"possible_ratings"} label={"Possible ratings"} multiple options={options}/>
}

type LineSelectProps = {
    PossibleRatings: string[],
    Index: number,
}

function LineSelect({PossibleRatings, Index}: LineSelectProps) {
    const {data, isFetching} = useGetRatingsQuery()
    if (isFetching) {
        return <Spin/>
    }

    const options = data?.resource?.map((r: Rating) => {
        const label = <RatingTooltip RatingId={r.id}/>
        return {label: label, value: r.id}
    })

    return <DoubleSelectFormItem dualLabel={`Line ${Index + 1} ratings`}
                                 name1={["lines", Index, "player_1_rating"]}
                                 name2={["lines", Index, "player_2_rating"]}
                                 options1={options} options2={options}/>

}
