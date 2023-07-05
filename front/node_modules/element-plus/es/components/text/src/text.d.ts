import type Text from './text.vue';
import type { ExtractPropTypes } from 'vue';
export declare const textProps: {
    readonly type: import("element-plus/es/utils").EpPropFinalized<StringConstructor, "" | "success" | "warning" | "info" | "primary" | "danger", unknown, "", boolean>;
    readonly size: import("element-plus/es/utils").EpPropFinalized<StringConstructor, "" | "default" | "small" | "large", unknown, "", boolean>;
    readonly truncated: {
        readonly type: import("vue").PropType<import("element-plus/es/utils").EpPropMergeType<BooleanConstructor, unknown, unknown>>;
        readonly required: false;
        readonly validator: ((val: unknown) => boolean) | undefined;
        __epPropKey: true;
    };
    readonly tag: import("element-plus/es/utils").EpPropFinalized<StringConstructor, unknown, unknown, "span", boolean>;
};
export declare type TextProps = ExtractPropTypes<typeof textProps>;
export declare type TextInstance = InstanceType<typeof Text>;
