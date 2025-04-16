import * as React from 'react';
import {InputFormItem, NumberInputFormItem} from "../../common/FormItem";
import {useCreateRatingMutation, useUpdateRatingMutation} from "../../redux/api.js";
import {CommonFormModal} from "../../common/CommonFormModal";

type RatingFormProps = {
    Update?: boolean // is this updating an existing record or creating a new record
    RatingId?: string // this will be provided on an update
    InitialState?: Rating // this will be provided on an update
}

export function RatingForm({Update, InitialState, RatingId}: RatingFormProps) {
    const [createRating] = useCreateRatingMutation()
    const [updateRating] = useUpdateRatingMutation()

    const onSubmit = async (formValues: any) => {
        const body: Rating = {
            name: formValues.name,
            description: formValues.description,
        }
        let func = () => createRating(body)
        if (Update) {
            func = () => updateRating({id: RatingId, body: body})
        }
        return func();
    }

    return <CommonFormModal ObjectType={"rating"} IsUpdate={Update} OnSubmit={onSubmit} InitialState={InitialState}>
        <InputFormItem name={"name"} label={"Name"}/>
        <InputFormItem name={"description"} label={"Description"}/>
    </CommonFormModal>
}
