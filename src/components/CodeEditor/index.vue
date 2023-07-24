<!-- <template>
  <div>
    <textarea ref="textareaRef" id="" cols="30" rows="10"></textarea>
  </div>
</template>

<script setup lang="ts">
import { ref, watchEffect, onMounted, onUnmounted } from "vue";
import CodeMirror from "codemirror";
import "codemirror/lib/codemirror.css";
import "codemirror/mode/css/css.js";
import "codemirror/mode/yaml/yaml.js";
import "codemirror/mode/yaml-frontmatter/yaml-frontmatter.js";
import "`codemirror`/mode/javascript/javascript";
import "codemirror/addon/selection/active-line"; // 代码高亮
import "codemirror/addon/fold/foldgutter.css";
import "codemirror/addon/fold/foldcode";
import "codemirror/addon/fold/brace-fold";
import "codemirror/addon/scroll/simplescrollbars.css";
import "codemirror/addon/scroll/simplescrollbars";
import "codemirror/addon/hint/show-hint";
import "codemirror/addon/hint/javascript-hint";
import "codemirror/addon/hint/anyword-hint";
import "codemirror/addon/hint/css-hint";
import "codemirror/addon/hint/show-hint.css";
import "codemirror/theme/rubyblue.css";
import "codemirror/theme/xq-light.css";
import "codemirror/theme/monokai.css"; // 主题

const props = defineProps({
  value: {
    type: [Object, String] as () => Object | String,
    default: null
  },
  mode: {
    type: String,
    default: "yaml" // yaml / application/json
  },
  readOnly: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(["change", "input"]);

const textareaRef = ref(null);
let codeMirrorEditor: CodeMirror.EditorFromTextArea | null = null;

const initCodeMirrorEditor = () => {
  codeMirrorEditor = CodeMirror.fromTextArea(textareaRef.value, {
    mode: props.mode,
    lineNumbers: true, // 显示行数
    lint: true,
    indentUnit: 2, // 缩进单位为2
    smartIndent: true, // 自动缩进是否开启
    styleActiveLine: true, // 当前行背景高亮
    matchBrackets: true, // 括号匹配
    lineWrapping: true, // 自动换行
    tabSize: 2,
    styleActiveLine: true, // 设置光标所在行高亮
    readOnly: props.readOnly,
    theme: "rubyblue",
    scrollbarStyle: "overlay"
  });

  if (props.value) codeMirrorEditor.setValue(props.value as string); // 设置值
  codeMirrorEditor.on("change", () => {
    const EditorValue = codeMirrorEditor.getValue();
    emit("change", EditorValue);
    emit("input", EditorValue);
  });
};

watchEffect(() => {
  if (!codeMirrorEditor) return;
  const editorValue = codeMirrorEditor.getValue();
  if (props.value !== editorValue) {
    codeMirrorEditor.setValue(props.value as string);
  }
});

onMounted(() => {
  setTimeout(initCodeMirrorEditor, 500);
});

onUnmounted(() => {
  if (codeMirrorEditor) {
    codeMirrorEditor.off("change");
  }
});
</script> -->
