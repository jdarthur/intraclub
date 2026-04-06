import {
    useCreateDraftMutation,
    useGetDraftOrderPatternsQuery,
    useGetFormatsQuery,
    useUpdateDraftMutation,
} from "../redux/api.js";
import {Form, Popover, Space, Spin, Tag} from "antd";
import {SubmitResult} from "../common/CommonFormModal";
import {InputFormItem, SelectFormItem} from "../common/FormItem";
import {StepFormStep} from "../common/StepForm";
import {StepFormModal} from "../common/StepFormModal";
import {RatingTooltip} from "../settings/formats/Formats";
import {InfoCircleOutlined} from '@ant-design/icons';
import * as React from "react";
import {TabbedPopover, TabbedPopoverTab} from "../common/TabbedPopover";

type DraftFormProps = {
    Update?: boolean // is this updating an existing record or creating a new record?
    DraftId?: string // this will be provided on an update
    InitialState?: Draft // this will be provided on an update
}

export function DraftForm({Update, InitialState, DraftId}: DraftFormProps) {
    const [createDraft] = useCreateDraftMutation()
    const [updateDraft] = useUpdateDraftMutation()

    const [f] = Form.useForm()

    const onSubmit = async (): Promise<SubmitResult> => {
        const values = f.getFieldsValue(true)
        const body: Draft = {
            name: values.name,
            captains: [values.captains],
            available: [],
            format: values.format,
            rating_cutoffs: values.rating_cutoffs,
            draft_order_pattern: values.draft_order_pattern,
        }
        let func = () => createDraft(body)
        if (Update) {
            func = () => updateDraft({id: DraftId, body: body})
        }
        return func();
    }

    const onChange = (changed: any, values: any) => {
        console.log("changed", changed)
    }

    const infoStep = <div>
        <InputFormItem name={"name"} label={"Name"}/>
        <FormatSelect/>
        <DraftOrderPatternSelect/>
    </div>


    const players = <div>
        Players...
    </div>

    const steps: StepFormStep[] = [
        {title: "Basic info", content: infoStep},
        {title: "Players", content: players},
    ]

    return <StepFormModal ObjectType={"season"} IsUpdate={Update}
                          InitialState={InitialState} form={f} onValuesChange={onChange} steps={steps}
                          onStepFormFinish={onSubmit} children={null} footer={null}/>
}

export function FormatSelect() {
    const {data, isFetching} = useGetFormatsQuery()
    if (isFetching) {
        return <Spin/>
    }

    const options = data?.resource?.map((f: Format) => {
        const label = <FormatTooltip possible_ratings={f.possible_ratings} name={f.name} lines={f.lines}/>
        return {label: label, value: f.id}
    })

    return <SelectFormItem name={"format"} label={"Format"} options={options}/>
}

function DraftOrderPatternSelect() {
    const {data, isFetching} = useGetDraftOrderPatternsQuery()

    if (isFetching) {
        return <Spin/>
    }

    const options = data?.resource?.map((d: DraftOrderPattern) => {
        const label = <DraftOrderPatternTooltip name={d.name} description={d.description} example={d.example}/>
        return {label: label, value: d.name}
    })

    return <SelectFormItem name={"draft_order_pattern"} label={"Draft order pattern"} options={options}/>
}

function FormatTooltip({possible_ratings, name, lines}: Format) {
    console.log("lines:", lines)
    const ratings = possible_ratings?.map((r: string) => <RatingTooltip RatingId={r}/>)
    const title = <div style={{display: "flex", flexDirection: "column"}}>
        {ratings}
    </div>
    return <Space size={"small"}>
        {name}
        <Popover title={"Format info"} content={title} trigger={"hover"} placement={"right"}>
            <InfoCircleOutlined/>
        </Popover>
    </Space>
}

function DraftOrderPatternTooltip({name, description, example}: DraftOrderPattern) {
    const tabs: TabbedPopoverTab[] = [
        {title: "Description", content: <i>{description}</i>},
        {title: "Example", content: <DraftOrderExample example={example} name={""} description={""}/>},
    ]

    const content = <TabbedPopover Tabs={tabs} style={{maxWidth: "20em"}}/>

    return <Space size={"small"}>
        {name}
        {content}
    </Space>
}

function DraftOrderExample({
                               example
                           }: DraftOrderPattern) {

    const body = example?.map((round: number[], roundIndex) => {

        const picks = round.map((teamID: number, index: number) => {
            const trailing = index === (round.length) - 1 ? "" : ", "
            return <span>{teamID}{trailing}</span>
        })

        const trailing2 = roundIndex === (example.length) - 1 ? "" : <span style={{marginRight: "0.5em"}}>➔</span>

        return <span>
            <Tag>{picks}</Tag>
            {trailing2}
        </span>

    })

    return <div style={{display: "flex", flexDirection: "row", flexWrap: "wrap", maxWidth: "15em"}}>
        {body}
    </div>
}