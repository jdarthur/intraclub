import * as React from "react"
import {Alert, Button, Space, Steps} from "antd";
import {useState} from "react";
import {SubmitResult} from "./CommonFormModal";

export type StepFormStep = {
    title: string
    content: any
    onNext?: () => Promise<SubmitResult>
}

export type StepFormProps = {
    steps: StepFormStep[]
    onStepFormFinish: () => Promise<SubmitResult>
    update?: boolean
    setDisabled?: (b: boolean) => void
    error?: string
}

export function StepForm({steps, onStepFormFinish, update, setDisabled, error}: StepFormProps) {

    const [currentStep, setCurrentStep] = useState<number>(0)
    const [_disabled, _setDisabled] = useState<boolean>(false)

    const SetDisabledActual = (b: boolean) => {
        if (setDisabled) {
            setDisabled(b)
        } else {
            _setDisabled(b)
        }
    }

    const visibleStep = steps[currentStep]
    const onLastStep = currentStep === (steps.length - 1)

    const back = () => {
        // if we hit back we will be guaranteed not on the last page
        // so we can set disabled == false in the child component
        SetDisabledActual(false)
        setCurrentStep(currentStep - 1)
    }

    const next = async () => {

        // if we are currently on the last step, we will
        // call the onFinish function which should save
        // the data and close the form
        if (onLastStep) {
            await onStepFormFinish()
            return
        }

        // otherwise, run any onNext logic if it exists
        if (visibleStep.onNext) {
            const res = await visibleStep.onNext()
            if (res.error) {
                console.log(res.error)
            }
        }

        if ((currentStep + 1) == (steps.length - 1)) {
            SetDisabledActual(true)
        }

        // increment the current step so we go to the next sub-page
        setCurrentStep(currentStep + 1)
    }

    const createOrUpdate = update ? "Update" : "Create"
    const nextButtonTitle = onLastStep ? createOrUpdate : "Next"

    const onChange = async (value: number) => {
        const isNowLastStep = value == (steps.length - 1)
        SetDisabledActual(isNowLastStep)

        if (visibleStep.onNext) {
            const res = await visibleStep.onNext()
            if (res.error) {
                console.log(res.error)
            }
        }

        setCurrentStep(value)
    };


    return <div>
        <Steps items={steps} current={currentStep} onChange={onChange}/>
        <div style={{marginBottom: "1.5em"}}/>
        <div style={{display: "flex", flexDirection: "column"}}>
            {visibleStep.content}

            {error? <Alert message={error} type={"error"} style={{marginBottom: "1.5em"}}/>:null}

            <Space style={{alignSelf: "flex-end"}}>
                <Button onClick={back} disabled={currentStep == 0}>
                    Back
                </Button>
                <Button onClick={next} type={"primary"}>
                    {nextButtonTitle}
                </Button>
            </Space>


        </div>
    </div>

}