import * as React from 'react';
import {InputFormItem, SelectFormItem,} from "../../common/FormItem";
import {useCreateFormatMutation, useGetRatingsQuery, useUpdateFormatMutation} from "../../redux/api.js";
import {CommonFormModal} from "../../common/CommonFormModal";
import {Space, Tooltip} from "antd";
import {InfoCircleOutlined} from "@ant-design/icons";

type FormatFormProps = {
    Update?: boolean // is this updating an existing record or creating a new record
    FormatId?: string // this will be provided on an update
    InitialState?: Format // this will be provided on an update
}

export function FormatForm({Update, InitialState, FormatId}: FormatFormProps) {
    const [createFormat] = useCreateFormatMutation()
    const [updateFormat] = useUpdateFormatMutation()

    const onSubmit = async (formValues: any) => {
        const body: Format = {
            name: formValues.name,
            possible_ratings: formValues.possible_ratings,
            lines: formValues.lines,
        }
        let func = () => createFormat(body)
        if (Update) {
            func = () => updateFormat({id: FormatId, body: body})
        }
        return func();
    }

    return <CommonFormModal ObjectType={"format"} IsUpdate={Update} OnSubmit={onSubmit} InitialState={InitialState}>
        <InputFormItem name={"name"} label={"Name"}/>
        <RatingSelect />
    </CommonFormModal>
}

function RatingSelect() {

    const {data} = useGetRatingsQuery()

    const options = data?.resource?.map((r: Rating) => {
        const label = <Space size={"small"}>
            {r.name}
            <Tooltip title={r.description}>
                <InfoCircleOutlined/>
            </Tooltip>
        </Space>
        return {label: label, value: r.id}
    })

    return <SelectFormItem name={"possible_ratings"} label={"Possible ratings"} multiple options={options}/>

}
