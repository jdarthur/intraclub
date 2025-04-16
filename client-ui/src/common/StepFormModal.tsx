import {StepForm, StepFormProps} from "./StepForm";
import * as React from "react";
import {useState} from "react";
import {CommonFormModal, CommonModalProps, SubmitResult} from "./CommonFormModal";


type StepFormModalProps = StepFormProps & CommonModalProps

export function StepFormModal({...props}: StepFormModalProps) {

    const [open, setOpen] = useState<boolean>(false)
    const [error, setError] = useState<string>("")

    const onSubmit = async (): Promise<SubmitResult> => {
        const res = await props.onStepFormFinish()
        if (res.error) {
            console.log("error on step form modal submit:", res.error)
            setError(res.error.data.error)
            return res
        }
        setOpen(false)
        return {data: null, error: null}
    }

    const onCancel = async (): Promise<SubmitResult> => {
        setOpen(false)
        return {data: null, error: null}
    }


    return <CommonFormModal open={open} ObjectType={props.ObjectType} IsUpdate={props.IsUpdate}
                            InitialState={props.InitialState} setOpen={setOpen} footer={props.footer}
                            form={props.form} onCancel={onCancel} onValuesChange={props.onValuesChange}>
        <StepForm steps={props.steps} onStepFormFinish={onSubmit}
                  update={props.IsUpdate} setDisabled={props.setDisabled}
                  error={error}/>
    </CommonFormModal>


}