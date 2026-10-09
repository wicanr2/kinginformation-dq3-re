"""Reversible C-subset ABI lowering prototype; entry: docs/25-match-progress.md.

Input is C plus instruction-free aux declarations. This module accepts no EXE,
addresses, original bytes or compiler object. Unsupported syntax fails closed.
The output is symbolic assembler and typed IR, not a patched machine-code blob.
"""

import re


REGISTERS = ("ax", "bx", "cx", "dx", "si", "di", "bp", "es")
TOKEN = re.compile(r"\s+|0[xX][0-9a-fA-F]+[uU]?|[0-9]+[uU]?|[A-Za-z_][A-Za-z0-9_]*|[(){};,=]")


class UnsupportedC(ValueError):
    pass


def integer(token):
    token = token.rstrip("uU")
    base = 16 if token.lower().startswith("0x") else 8 if len(token) > 1 and token.startswith("0") else 10
    try:
        value = int(token, base)
    except ValueError as error:
        raise UnsupportedC("Invalid C integer constant") from error
    if not 0 <= value <= 65535:
        raise UnsupportedC("Constant is outside the unsigned 16-bit target")
    return value


def parse_abi(line):
    match = re.fullmatch(r'#pragma\s+aux\s+(\w+)\s+"_\*"\s+(.*?)\s*;', line.strip())
    if not match or "=" in line:
        raise UnsupportedC("Only instruction-free aux declarations are supported")
    name, tail = match.groups()
    abi = {"far": False, "params": [], "returns": [], "modifies": None}
    if tail.startswith("far "):
        abi["far"], tail = True, tail[4:]
    for keyword, field in [("parm", "params"), ("value", "returns")]:
        if tail.startswith(keyword + " "):
            tail = tail[len(keyword):].lstrip()
            while tail.startswith("["):
                end = tail.find("]")
                if end == -1:
                    raise UnsupportedC("Unclosed ABI register set")
                registers = tail[1:end].split()
                if len(registers) != 1 or registers[0] not in REGISTERS:
                    raise UnsupportedC("Only scalar register arguments/results are supported")
                abi[field].append(registers[0])
                tail = tail[end + 1:].lstrip()
    match = re.fullmatch(r"modify\s+exact\s+\[([a-z ]*)\]", tail)
    if not match:
        raise UnsupportedC("An exact, explicit clobber declaration is required")
    modifies = match.group(1).split()
    if len(set(modifies)) != len(modifies) or any(r not in REGISTERS for r in modifies):
        raise UnsupportedC("Unknown or duplicate clobber register")
    abi["modifies"] = modifies
    return name, abi


class Parser:
    def __init__(self, source):
        # This subset has no string literals outside aux declarations.
        source = re.sub(r"/\*.*?\*/|//[^\n]*", "", source, flags=re.S)
        self.abis, lines = {}, []
        for line in source.splitlines():
            if line.lstrip().startswith("#"):
                name, abi = parse_abi(line)
                if name in self.abis:
                    raise UnsupportedC("Duplicate ABI declaration")
                self.abis[name] = abi
            else:
                lines.append(line)
        source = "\n".join(lines)
        self.tokens, position = [], 0
        while position < len(source):
            match = TOKEN.match(source, position)
            if not match:
                raise UnsupportedC("Unsupported C token near " + source[position:position + 30])
            token = match.group()
            if not token.isspace():
                self.tokens.append(token)
            position = match.end()
        self.position, self.globals, self.functions = 0, {}, {}

    def peek(self):
        return self.tokens[self.position] if self.position < len(self.tokens) else None

    def take(self, expected=None):
        token = self.peek()
        if token is None or expected is not None and token != expected:
            raise UnsupportedC("Expected " + str(expected) + ", got " + str(token))
        self.position += 1
        return token

    def identifier(self):
        token = self.take()
        if not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", token):
            raise UnsupportedC("Expected identifier")
        if token in ("__asm", "_asm", "asm"):
            raise UnsupportedC("Assembly is not a C input")
        if token in ("auto", "break", "case", "char", "const", "continue", "default", "do", "double",
                     "else", "enum", "extern", "float", "for", "goto", "if", "int", "long", "register",
                     "return", "short", "signed", "sizeof", "static", "struct", "switch", "typedef",
                     "union", "unsigned", "void", "volatile", "while"):
            raise UnsupportedC("Keyword cannot be an identifier")
        return token

    def ctype(self):
        token = self.take()
        if token not in ("unsigned", "void"):
            raise UnsupportedC("Only target u16 and void types are supported")
        return "u16" if token == "unsigned" else "void"

    def expression(self):
        token = self.peek()
        if token is not None and token[0].isdigit():
            return {"op": "constant", "value": integer(self.take())}
        name = self.identifier()
        if self.peek() != "(":
            return {"op": "name", "name": name}
        self.take("(")
        args = []
        if self.peek() != ")":
            while True:
                args.append(self.expression())
                if self.peek() != ",":
                    break
                self.take(",")
        self.take(")")
        return {"op": "call", "function": name, "args": args}

    def body(self):
        self.take("{")
        statements = []
        while self.peek() != "}":
            if len(statements) >= 12:
                raise UnsupportedC("Prototype statement budget exceeded")
            if self.peek() == "unsigned":
                self.take()
                name = self.identifier()
                value = None
                if self.peek() == "=":
                    self.take()
                    value = self.expression()
                statements.append({"op": "declare", "name": name, "value": value})
            elif self.peek() == "return":
                self.take()
                statements.append({"op": "return", "value": self.expression()})
            else:
                value = self.expression()
                if self.peek() == "=":
                    self.take()
                    if value["op"] != "name":
                        raise UnsupportedC("Only named scalar assignment is supported")
                    statements.append({"op": "assign", "name": value["name"], "value": self.expression()})
                elif value["op"] == "call":
                    statements.append(value)
                else:
                    raise UnsupportedC("Unsupported expression statement")
            self.take(";")
        self.take("}")
        return statements

    def parse(self):
        while self.peek() is not None:
            external = self.peek() == "extern"
            if external:
                self.take()
            volatile = self.peek() == "volatile"
            if volatile:
                self.take()
            ctype, name = self.ctype(), self.identifier()
            if self.peek() != "(":
                if not external or not volatile or ctype != "u16" or name in self.globals or name in self.functions:
                    raise UnsupportedC("Globals must be distinct extern volatile u16 objects")
                self.take(";")
                self.globals[name] = "volatile-u16"
                continue
            if volatile:
                raise UnsupportedC("Volatile function type is unsupported")
            if name in self.globals:
                raise UnsupportedC("Function conflicts with a global object")
            self.take("(")
            params = []
            if self.peek() == "void":
                self.take()
            elif self.peek() != ")":
                while True:
                    if self.ctype() != "u16":
                        raise UnsupportedC("Only u16 parameters are supported")
                    params.append(self.identifier())
                    if self.peek() != ",":
                        break
                    self.take(",")
            self.take(")")
            definition = self.peek() == "{"
            body = self.body() if definition else None
            if not definition:
                self.take(";")
            previous = self.functions.get(name)
            if previous and (previous["type"] != ctype or previous["params"] != params or previous["body"] is not None):
                raise UnsupportedC("Inconsistent or duplicate function declaration")
            self.functions[name] = {"type": ctype, "params": params, "body": body,
                                    "visible_globals": list(self.globals),
                                    "visible_functions": list(self.functions)}
        definitions = [n for n, f in self.functions.items() if f["body"] is not None]
        if len(definitions) != 1 or set(self.abis) != set(self.functions):
            raise UnsupportedC("One definition and explicit ABI for every function are required")
        for name, function in self.functions.items():
            abi = self.abis[name]
            if len(abi["params"]) != len(function["params"]):
                raise UnsupportedC("Parameter/ABI arity differs")
            if function["type"] == "void" and abi["returns"] or function["type"] == "u16" and abi["returns"] != ["ax"]:
                raise UnsupportedC("This prototype supports only void or AX return")
            if function["type"] == "u16" and "ax" not in abi["modifies"]:
                raise UnsupportedC("This subset requires explicit AX clobber for an AX result")
        return definitions[0]


def name_value(value, name):
    return value == {"op": "name", "name": name}


def compile_source(source, profile):
    parser = Parser(source)
    entry = parser.parse()
    function, abi = parser.functions[entry], parser.abis[entry]
    if abi["far"] or function["params"]:
        raise UnsupportedC("Prototype entries are near and have no input parameters")
    statements = function["body"]
    visible = set(function["visible_globals"]) | set(function["params"])
    def check_expression(expression):
        if expression is None or expression["op"] == "constant":
            return
        if expression["op"] == "name":
            if expression["name"] not in visible:
                raise UnsupportedC("Value used before declaration")
        elif expression["op"] == "call":
            if expression["function"] not in function["visible_functions"]:
                raise UnsupportedC("Called function must be declared before the definition")
            for argument in expression["args"]:
                check_expression(argument)
    for statement in statements:
        if statement["op"] == "declare":
            check_expression(statement["value"])
            visible.add(statement["name"])
        elif statement["op"] == "assign":
            if statement["name"] not in visible:
                raise UnsupportedC("Assignment before declaration")
            check_expression(statement["value"])
        elif statement["op"] == "return":
            check_expression(statement["value"])
        else:
            check_expression(statement)
    locals_ = [statement["name"] for statement in statements if statement["op"] == "declare"]
    if len(set(locals_)) != len(locals_) or any(n in parser.globals or n in parser.functions for n in locals_):
        raise UnsupportedC("Duplicate or shadowing local declaration")
    returned = None
    if profile == "scoped-word-restore-v1":
        uninitialized = [statement["name"] for statement in statements
                         if statement["op"] == "declare" and statement["value"] is None]
        if len(uninitialized) > 1:
            raise UnsupportedC("Only the optional captured result may be uninitialized")
        statements = [statement for statement in statements
                      if not (statement["op"] == "declare" and statement["value"] is None)]
        if not 4 <= len(statements) <= 6:
            raise UnsupportedC("Not a bounded save/set/call/restore transaction")
        saved, setting = statements[:2]
        if (saved["op"] != "declare" or saved["value"] is None or saved["value"]["op"] != "name"
                or setting["op"] != "assign" or setting["value"]["op"] != "constant"):
            raise UnsupportedC("Expected snapshot and constant write")
        global_name = saved["value"]["name"]
        if global_name not in parser.globals or setting["name"] != global_name:
            raise UnsupportedC("Snapshot and write must reference one volatile u16 global")
        tail = statements[2:]
        if uninitialized:
            returned = uninitialized[0]
        call = tail.pop(0)
        if returned:
            if call["op"] != "assign" or call["name"] != returned or call["value"]["op"] != "call":
                raise UnsupportedC("Expected captured AX call result")
            call = call["value"]
        if not tail or tail[0] != {"op": "assign", "name": global_name, "value": {"op": "name", "name": saved["name"]}}:
            raise UnsupportedC("The original snapshot must be restored after the call")
        tail.pop(0)
        if returned:
            if tail != [{"op": "return", "value": {"op": "name", "name": returned}}] or function["type"] != "u16":
                raise UnsupportedC("Only the captured AX result may be returned")
        elif tail or function["type"] != "void":
            raise UnsupportedC("Unsupported transaction tail")
        ir = {"profile": profile, "snapshot": global_name, "write": setting["value"]["value"],
              "call": call, "restore": global_name, "return_AX": bool(returned)}
    elif profile == "exact-gpr-envelope-v1":
        if len(statements) != 1 or statements[0]["op"] != "call" or function["type"] != "void":
            raise UnsupportedC("Envelope must contain exactly one void call")
        call = statements[0]
        ir = {"profile": profile, "call": call}
    else:
        raise UnsupportedC("Unknown opt-in ABI profile")
    if call["op"] != "call" or call["function"] == entry or call["function"] not in parser.functions:
        raise UnsupportedC("Expected one declared external function")
    callee, callee_abi = parser.functions[call["function"]], parser.abis[call["function"]]
    if callee["body"] is not None or not callee_abi["far"] or len(call["args"]) != len(callee["params"]):
        raise UnsupportedC("Expected a far call with explicit register arguments")
    if returned and (callee["type"] != "u16" or callee_abi["returns"] != ["ax"]):
        raise UnsupportedC("Captured call must return AX")
    if any(x["op"] != "constant" for x in call["args"]):
        raise UnsupportedC("Only scalar constant arguments are supported")
    if len(call["args"]) > 1:
        raise UnsupportedC("This bounded prototype supports zero or one argument")
    if any(r not in ("ax", "bx", "cx", "dx", "si", "di") for r in callee_abi["params"]):
        raise UnsupportedC("Unsupported scalar argument register")
    changed = set(callee_abi["modifies"]) | set(callee_abi["params"]) | set(callee_abi["returns"])
    preserved = changed - set(abi["modifies"]) - set(abi["returns"])
    saves = [register for register in REGISTERS if register in preserved]
    ir.update(entry=entry, callee=call["function"], saved_registers=saves,
              argument_registers=callee_abi["params"], ABI_policy="opt-in exact preservation includes void AX and BP")
    symbol = lambda name: "_" + name
    lines = [".8086", "PUBLIC " + symbol(entry), "EXTRN " + symbol(call["function"]) + ":FAR"]
    if profile == "scoped-word-restore-v1":
        lines.append("EXTRN " + symbol(ir["snapshot"]) + ":WORD")
    lines += ["_TEXT SEGMENT BYTE PUBLIC USE16 'CODE'", "ASSUME CS:_TEXT, DS:NOTHING, SS:NOTHING", symbol(entry) + ":"]
    lines += ["    push " + register for register in saves]
    if profile == "scoped-word-restore-v1":
        lines += ["    push word ptr " + symbol(ir["snapshot"]),
                  "    mov word ptr " + symbol(ir["snapshot"]) + ",0%Xh" % ir["write"]]
    lines += ["    mov " + register + ",0%Xh" % arg["value"] for register, arg in zip(callee_abi["params"], call["args"])]
    lines.append("    call far ptr " + symbol(call["function"]))
    if profile == "scoped-word-restore-v1":
        lines.append("    pop word ptr " + symbol(ir["restore"]))
    lines += ["    pop " + register for register in reversed(saves)]
    lines += ["    ret", "_TEXT ENDS", "END"]
    return {"IR": ir, "assembly": "\n".join(lines) + "\n", "public_symbol": symbol(entry),
            "input_language": "restricted C u16/void subset plus instruction-free ABI declarations",
            "optimizer_reads_original_EXE": False, "formal_C_coverage_increment": 0}
