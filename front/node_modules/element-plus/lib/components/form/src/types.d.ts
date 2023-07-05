import type { SetupContext, UnwrapRef } from 'vue';
import type { RuleItem, ValidateError, ValidateFieldsError } from 'async-validator';
import type { ComponentSize } from 'element-plus/es/constants';
import type { Arrayable } from 'element-plus/es/utils';
import type { FormItemProp, FormItemProps, FormItemValidateState } from './form-item';
import type { FormEmits, FormProps } from './form';
import type { useFormLabelWidth } from './utils';
export declare type FormLabelWidthContext = ReturnType<typeof useFormLabelWidth>;
export interface FormItemRule extends RuleItem {
    trigger?: Arrayable<string>;
}
export declare type FormRules = Partial<Record<string, Arrayable<FormItemRule>>>;
export declare type FormValidationResult = Promise<boolean>;
export declare type FormValidateCallback = (isValid: boolean, invalidFields?: ValidateFieldsError) => void;
export interface FormValidateFailure {
    errors: ValidateError[] | null;
    fields: ValidateFieldsError;
}
export declare type FormContext = FormProps & UnwrapRef<FormLabelWidthContext> & {
    emit: SetupContext<FormEmits>['emit'];
    addField: (field: FormItemContext) => void;
    removeField: (field: FormItemContext) => void;
    resetFields: (props?: Arrayable<FormItemProp>) => void;
    clearValidate: (props?: Arrayable<FormItemProp>) => void;
    validateField: (props?: Arrayable<FormItemProp>, callback?: FormValidateCallback) => FormValidationResult;
};
export interface FormItemContext extends FormItemProps {
    $el: HTMLDivElement | undefined;
    size: ComponentSize;
    validateState: FormItemValidateState;
    isGroup: boolean;
    labelId: string;
    inputIds: string[];
    hasLabel: boolean;
    addInputId: (id: string) => void;
    removeInputId: (id: string) => void;
    validate: (trigger: string, callback?: FormValidateCallback) => FormValidationResult;
    resetField(): void;
    clearValidate(): void;
}
