from pathlib import Path

from docx import Document
from docx.enum.section import WD_SECTION
from docx.enum.table import WD_ALIGN_VERTICAL, WD_TABLE_ALIGNMENT
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
from docx.shared import Inches, Pt, RGBColor


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "USER_REDEEM_AND_API_GUIDE_CN.docx"

BLUE = RGBColor(46, 116, 181)
DARK_BLUE = RGBColor(31, 77, 120)
INK = RGBColor(33, 37, 41)
MUTED = RGBColor(91, 103, 112)
LIGHT_BLUE = "E8EEF5"
LIGHT_GRAY = "F2F4F7"
CALLOUT_FILL = "F4F6F9"
BORDER = "D7DBE2"


def set_cell_shading(cell, fill):
    tc_pr = cell._tc.get_or_add_tcPr()
    shd = tc_pr.find(qn("w:shd"))
    if shd is None:
        shd = OxmlElement("w:shd")
        tc_pr.append(shd)
    shd.set(qn("w:fill"), fill)


def set_cell_margins(cell, top=100, start=140, bottom=100, end=140):
    tc_pr = cell._tc.get_or_add_tcPr()
    tc_mar = tc_pr.find(qn("w:tcMar"))
    if tc_mar is None:
        tc_mar = OxmlElement("w:tcMar")
        tc_pr.append(tc_mar)
    for margin_name, value in {
        "top": top,
        "start": start,
        "bottom": bottom,
        "end": end,
    }.items():
        node = tc_mar.find(qn(f"w:{margin_name}"))
        if node is None:
            node = OxmlElement(f"w:{margin_name}")
            tc_mar.append(node)
        node.set(qn("w:w"), str(value))
        node.set(qn("w:type"), "dxa")


def set_table_borders(table, color=BORDER, size="6"):
    tbl_pr = table._tbl.tblPr
    borders = tbl_pr.find(qn("w:tblBorders"))
    if borders is None:
        borders = OxmlElement("w:tblBorders")
        tbl_pr.append(borders)
    for edge in ["top", "left", "bottom", "right", "insideH", "insideV"]:
        tag = f"w:{edge}"
        element = borders.find(qn(tag))
        if element is None:
            element = OxmlElement(tag)
            borders.append(element)
        element.set(qn("w:val"), "single")
        element.set(qn("w:sz"), size)
        element.set(qn("w:space"), "0")
        element.set(qn("w:color"), color)


def set_table_geometry(table, widths):
    table.alignment = WD_TABLE_ALIGNMENT.LEFT
    table.autofit = False
    tbl_pr = table._tbl.tblPr
    tbl_w = tbl_pr.find(qn("w:tblW"))
    if tbl_w is None:
        tbl_w = OxmlElement("w:tblW")
        tbl_pr.append(tbl_w)
    total = sum(widths)
    tbl_w.set(qn("w:w"), str(total))
    tbl_w.set(qn("w:type"), "dxa")

    tbl_ind = tbl_pr.find(qn("w:tblInd"))
    if tbl_ind is None:
        tbl_ind = OxmlElement("w:tblInd")
        tbl_pr.append(tbl_ind)
    tbl_ind.set(qn("w:w"), "120")
    tbl_ind.set(qn("w:type"), "dxa")

    tbl_grid = table._tbl.tblGrid
    for child in list(tbl_grid):
        tbl_grid.remove(child)
    for width in widths:
        col = OxmlElement("w:gridCol")
        col.set(qn("w:w"), str(width))
        tbl_grid.append(col)

    for row in table.rows:
        for idx, cell in enumerate(row.cells):
            width = widths[min(idx, len(widths) - 1)]
            cell.width = Inches(width / 1440)
            tc_pr = cell._tc.get_or_add_tcPr()
            tc_w = tc_pr.find(qn("w:tcW"))
            if tc_w is None:
                tc_w = OxmlElement("w:tcW")
                tc_pr.append(tc_w)
            tc_w.set(qn("w:w"), str(width))
            tc_w.set(qn("w:type"), "dxa")
            cell.vertical_alignment = WD_ALIGN_VERTICAL.CENTER
            set_cell_margins(cell)


def set_run_font(run, size=None, color=None, bold=None, italic=None):
    run.font.name = "Calibri"
    run._element.rPr.rFonts.set(qn("w:ascii"), "Calibri")
    run._element.rPr.rFonts.set(qn("w:hAnsi"), "Calibri")
    run._element.rPr.rFonts.set(qn("w:eastAsia"), "Microsoft YaHei")
    if size is not None:
        run.font.size = Pt(size)
    if color is not None:
        run.font.color.rgb = color
    if bold is not None:
        run.bold = bold
    if italic is not None:
        run.italic = italic


def set_paragraph_spacing(paragraph, before=0, after=6, line=1.25):
    paragraph.paragraph_format.space_before = Pt(before)
    paragraph.paragraph_format.space_after = Pt(after)
    paragraph.paragraph_format.line_spacing = line


def add_heading(doc, text, level=1):
    p = doc.add_paragraph(style=f"Heading {level}")
    set_paragraph_spacing(
        p,
        before={1: 18, 2: 14, 3: 10}.get(level, 8),
        after={1: 10, 2: 7, 3: 5}.get(level, 5),
        line=1.1,
    )
    run = p.add_run(text)
    set_run_font(
        run,
        size={1: 16, 2: 13, 3: 12}.get(level, 12),
        color=BLUE if level < 3 else DARK_BLUE,
        bold=True,
    )
    return p


def add_para(doc, text="", bold_prefix=None, after=6):
    p = doc.add_paragraph()
    set_paragraph_spacing(p, after=after)
    if bold_prefix and text.startswith(bold_prefix):
        r1 = p.add_run(bold_prefix)
        set_run_font(r1, size=11, color=INK, bold=True)
        r2 = p.add_run(text[len(bold_prefix) :])
        set_run_font(r2, size=11, color=INK)
    else:
        r = p.add_run(text)
        set_run_font(r, size=11, color=INK)
    return p


def add_bullet(doc, text):
    p = doc.add_paragraph(style="List Bullet")
    set_paragraph_spacing(p, after=4)
    r = p.add_run(text)
    set_run_font(r, size=11, color=INK)
    return p


def add_number(doc, text):
    p = doc.add_paragraph(style="List Number")
    set_paragraph_spacing(p, after=4)
    r = p.add_run(text)
    set_run_font(r, size=11, color=INK)
    return p


def add_code(doc, text):
    p = doc.add_paragraph()
    set_paragraph_spacing(p, after=6, line=1.15)
    run = p.add_run(text)
    run.font.name = "Consolas"
    run._element.rPr.rFonts.set(qn("w:ascii"), "Consolas")
    run._element.rPr.rFonts.set(qn("w:hAnsi"), "Consolas")
    run._element.rPr.rFonts.set(qn("w:eastAsia"), "Microsoft YaHei")
    run.font.size = Pt(10)
    run.font.color.rgb = RGBColor(30, 48, 80)
    return p


def add_callout(doc, title, body):
    table = doc.add_table(rows=1, cols=1)
    set_table_geometry(table, [9360])
    set_table_borders(table, color="E2E7EE", size="4")
    cell = table.cell(0, 0)
    set_cell_shading(cell, CALLOUT_FILL)
    p = cell.paragraphs[0]
    set_paragraph_spacing(p, after=2)
    title_run = p.add_run(title)
    set_run_font(title_run, size=10.5, color=DARK_BLUE, bold=True)
    body_p = cell.add_paragraph()
    set_paragraph_spacing(body_p, after=0, line=1.2)
    body_run = body_p.add_run(body)
    set_run_font(body_run, size=10.5, color=INK)
    doc.add_paragraph().paragraph_format.space_after = Pt(2)
    return table


def add_table(doc, headers, rows, widths, header_fill=LIGHT_BLUE):
    table = doc.add_table(rows=1, cols=len(headers))
    set_table_geometry(table, widths)
    set_table_borders(table)
    hdr = table.rows[0]
    hdr._tr.get_or_add_trPr().append(OxmlElement("w:tblHeader"))
    for idx, header in enumerate(headers):
        cell = hdr.cells[idx]
        set_cell_shading(cell, header_fill)
        p = cell.paragraphs[0]
        set_paragraph_spacing(p, after=0, line=1.15)
        r = p.add_run(header)
        set_run_font(r, size=10.5, color=INK, bold=True)
    for row in rows:
        cells = table.add_row().cells
        for idx, value in enumerate(row):
            p = cells[idx].paragraphs[0]
            set_paragraph_spacing(p, after=0, line=1.15)
            r = p.add_run(value)
            set_run_font(r, size=10.5, color=INK)
    set_table_geometry(table, widths)
    doc.add_paragraph().paragraph_format.space_after = Pt(4)
    return table


def configure_styles(doc):
    styles = doc.styles
    normal = styles["Normal"]
    normal.font.name = "Calibri"
    normal.font.size = Pt(11)
    normal._element.rPr.rFonts.set(qn("w:eastAsia"), "Microsoft YaHei")
    normal.paragraph_format.space_after = Pt(6)
    normal.paragraph_format.line_spacing = 1.25

    for style_name in ["Heading 1", "Heading 2", "Heading 3"]:
        style = styles[style_name]
        style.font.name = "Calibri"
        style._element.rPr.rFonts.set(qn("w:eastAsia"), "Microsoft YaHei")
        style.font.bold = True
        style.paragraph_format.keep_with_next = True
        style.paragraph_format.line_spacing = 1.1

    styles["Heading 1"].font.size = Pt(16)
    styles["Heading 1"].font.color.rgb = BLUE
    styles["Heading 2"].font.size = Pt(13)
    styles["Heading 2"].font.color.rgb = BLUE
    styles["Heading 3"].font.size = Pt(12)
    styles["Heading 3"].font.color.rgb = DARK_BLUE

    for style_name in ["List Bullet", "List Number"]:
        style = styles[style_name]
        style.font.name = "Calibri"
        style.font.size = Pt(11)
        style._element.rPr.rFonts.set(qn("w:eastAsia"), "Microsoft YaHei")
        style.paragraph_format.left_indent = Inches(0.375)
        style.paragraph_format.first_line_indent = Inches(-0.188)
        style.paragraph_format.space_after = Pt(4)
        style.paragraph_format.line_spacing = 1.25


def add_header_footer(doc):
    section = doc.sections[0]
    section.top_margin = Inches(1)
    section.right_margin = Inches(1)
    section.bottom_margin = Inches(1)
    section.left_margin = Inches(1)
    section.header_distance = Inches(0.492)
    section.footer_distance = Inches(0.492)

    header = section.header
    p = header.paragraphs[0]
    p.alignment = WD_ALIGN_PARAGRAPH.RIGHT
    r = p.add_run("中转站新手说明书")
    set_run_font(r, size=9, color=MUTED)

    footer = section.footer
    p = footer.paragraphs[0]
    p.alignment = WD_ALIGN_PARAGRAPH.RIGHT
    r = p.add_run("请勿公开 API Key")
    set_run_font(r, size=9, color=MUTED)


def add_cover(doc):
    p = doc.add_paragraph()
    set_paragraph_spacing(p, before=10, after=0)
    kicker = p.add_run("新手上手指南")
    set_run_font(kicker, size=11, color=BLUE, bold=True)

    title = doc.add_paragraph()
    set_paragraph_spacing(title, before=3, after=6, line=1.05)
    r = title.add_run("中转站账号注册、兑换码与 API 使用说明书")
    set_run_font(r, size=24, color=DARK_BLUE, bold=True)

    subtitle = doc.add_paragraph()
    set_paragraph_spacing(subtitle, after=14, line=1.2)
    r = subtitle.add_run("面向第一次使用大模型 API 的用户：按步骤完成注册、兑换额度、复制 URL 和 Key，并在工具或代码中完成调用。")
    set_run_font(r, size=12.5, color=MUTED)

    add_table(
        doc,
        ["你会完成", "结果"],
        [
            ("注册账号", "能登录中转站后台"),
            ("兑换额度", "余额中能看到可用额度"),
            ("获取 URL 和 Key", "拿到 Base URL 与 API Key"),
            ("测试调用", "能成功收到大模型回复"),
        ],
        [2300, 7060],
        header_fill=LIGHT_GRAY,
    )

    add_callout(
        doc,
        "重要提醒",
        "本文中的域名、API Key 和模型名都是示例。实际填写时，请以平台后台展示的信息为准。API Key 等同于额度使用凭证，请像密码一样保管。",
    )


def build_doc():
    doc = Document()
    configure_styles(doc)
    add_header_footer(doc)
    add_cover(doc)

    add_heading(doc, "一分钟快速上手", 1)
    quick_steps = [
        "打开中转站网址，例如：https://你的中转站域名",
        "注册账号并登录后台。",
        "进入「兑换码」「充值」或「余额」页面，输入兑换码并确认兑换。",
        "进入「API Key」「令牌」或「密钥」页面，创建并复制 API Key。",
        "找到平台提供的 API URL，通常类似：https://你的中转站域名/v1",
        "在工具里填写 Base URL、API Key 和模型名。",
        "发送一句「你好」，能收到回复就说明配置成功。",
    ]
    for step in quick_steps:
        add_number(doc, step)

    add_heading(doc, "第一步：注册账号", 1)
    for step in [
        "打开中转站官网或客服提供的注册链接。",
        "点击「注册」「Sign up」或「创建账号」。",
        "按页面要求填写邮箱、手机号或用户名。",
        "设置密码，建议不低于 8 位，并混合字母、数字和符号。",
        "如果页面要求验证码，请按提示完成邮箱验证、短信验证或图形验证。",
        "注册成功后，回到登录页输入账号和密码登录。",
    ]:
        add_number(doc, step)
    add_para(doc, "注册完成后，建议先确认后台能看到「余额」「兑换码」和「API Key」等入口。")

    add_heading(doc, "第二步：使用兑换码兑换额度", 1)
    add_para(doc, "兑换码用于把平台赠送或购买的额度充入你的账号。")
    for step in [
        "登录中转站后台。",
        "找到「兑换码」「礼品码」「充值」「余额」或类似入口。",
        "将兑换码完整复制进去。",
        "点击「兑换」「确认」或「提交」。",
        "回到「余额」或「用量」页面，确认额度是否到账。",
    ]:
        add_number(doc, step)
    add_callout(
        doc,
        "兑换注意事项",
        "兑换码可能区分大小写；不要多复制空格；每个兑换码通常只能使用一次。如果提示已使用、已过期或不存在，请截图联系平台客服。",
    )

    add_heading(doc, "第三步：理解 URL 和 Key", 1)
    add_para(doc, "调用大模型时，最常用的是两个信息：URL 和 Key。")
    add_table(
        doc,
        ["名称", "你需要填写什么", "作用"],
        [
            ("API URL / Base URL", "通常类似 https://你的中转站域名/v1", "告诉工具把请求发到哪里"),
            ("API Key / Token / 密钥", "通常类似 sk-xxxxxxxx", "证明这次请求属于你的账号"),
            ("Model / 模型", "例如 gpt-5.5，或后台模型列表里的其他模型", "告诉平台你想调用哪个模型"),
        ],
        [2300, 3900, 3160],
    )
    for item in [
        "不要把 API Key 发到微信群、论坛、公开代码仓库或截图里。",
        "如果怀疑 Key 泄露，立刻在后台删除旧 Key，并创建新 Key。",
        "团队多人使用时，建议给不同成员或工具创建不同 Key。",
    ]:
        add_bullet(doc, item)

    add_heading(doc, "第四步：创建并复制 API Key", 1)
    for step in [
        "进入「API Key」「令牌」「密钥」或「开发者」页面。",
        "点击「创建 Key」「新建密钥」或类似按钮。",
        "给 Key 起一个容易识别的名字，例如 my-laptop、cursor-work 或 server-test。",
        "复制新生成的 API Key，并保存到安全位置。",
        "如果后台只在创建时显示一次 Key，请务必当场复制保存。",
    ]:
        add_number(doc, step)

    add_heading(doc, "第五步：在工具里填写 URL 和 Key", 1)
    add_para(doc, "大多数支持 OpenAI 兼容接口的工具都会有类似配置项：API Key、Base URL、Endpoint 和 Model。")
    add_table(
        doc,
        ["配置项", "示例填写"],
        [
            ("Base URL / API URL", "https://你的中转站域名/v1"),
            ("API Key", "sk-你的APIKey"),
            ("Model", "gpt-5.5 或后台模型列表中的模型 ID"),
        ],
        [2600, 6760],
    )
    add_para(doc, "如果工具要求填写完整接口地址，可以填写：")
    add_code(doc, "https://你的中转站域名/v1/chat/completions")
    add_para(doc, "如果工具要求填写 Base URL，通常只填写到 /v1：")
    add_code(doc, "https://你的中转站域名/v1")

    add_heading(doc, "PowerShell 测试示例", 1)
    add_para(doc, "把下面命令里的域名、Key 和模型名换成你自己的：")
    add_code(
        doc,
        '$body = @{ model = "gpt-5.5"; messages = @(@{ role = "user"; content = "你好，请用一句话介绍你自己。" }) } | ConvertTo-Json -Depth 5',
    )
    add_code(
        doc,
        'Invoke-RestMethod -Uri "https://你的中转站域名/v1/chat/completions" -Method Post -ContentType "application/json" -Headers @{ Authorization = "Bearer sk-你的APIKey" } -Body $body',
    )

    add_heading(doc, "Python 调用示例", 1)
    add_code(doc, "from openai import OpenAI")
    add_code(doc, 'client = OpenAI(api_key="sk-你的APIKey", base_url="https://你的中转站域名/v1")')
    add_code(
        doc,
        'response = client.chat.completions.create(model="gpt-5.5", messages=[{"role": "user", "content": "你好，请用一句话介绍你自己。"}])',
    )
    add_code(doc, "print(response.choices[0].message.content)")

    add_heading(doc, "使用中转站的优势", 1)
    advantages = [
        ("不用自己折腾 VPN 或海外网络", "很多用户直接访问海外大模型服务时，会遇到网络不稳定、连接失败或配置复杂的问题。使用中转站后，你只需要访问平台提供的 API URL，通常不需要自己单独准备 VPN 或海外网络环境。"),
        ("一个入口调用多种大模型", "中转站可以把多个模型入口集中到一个后台里。你可以按需使用 OpenAI 的 ChatGPT-5.5 / GPT-5.5 系列，以及平台支持的其他模型。实际开放哪些模型、模型 ID 怎么写、价格是多少，请以后台模型列表为准。"),
        ("对 OpenAI 兼容工具更友好", "很多开发工具、聊天客户端、IDE 插件和自动化工具都支持 OpenAI 兼容接口。只要工具允许填写自定义 Base URL 和 API Key，通常就可以接入中转站。"),
        ("余额、用量和成本更清楚", "后台通常可以查看余额、充值记录、兑换记录和调用用量。这样你能知道额度花在哪里，也能更容易控制成本。"),
        ("更适合团队和多工具使用", "你可以为不同工具或成员创建不同 API Key。后续如果某个 Key 不再使用，可以单独删除，不影响其他 Key。"),
        ("降低账号和接口维护成本", "用户不用分别研究多个模型厂商的账号、充值、接口格式和网络配置。中转站会把常用能力集中起来，让新用户更容易上手。"),
    ]
    for title, body in advantages:
        add_heading(doc, title, 2)
        add_para(doc, body)

    add_heading(doc, "常见问题", 1)
    faqs = [
        ("兑换码输入后提示无效怎么办？", "先检查是否复制错字符、是否多了空格、是否大小写错误。如果仍然无效，截图联系平台客服。"),
        ("为什么提示余额不足？", "可能是兑换码没有到账、余额已经用完，或当前模型单价较高。请先检查后台余额和用量记录。"),
        ("为什么提示 Key 错误或 401？", "通常是 API Key 复制错了、Key 已删除、Key 前后有空格，或者请求头没有写成 Authorization: Bearer sk-你的APIKey。"),
        ("为什么提示模型不存在？", "模型名必须和后台模型列表一致。请进入模型列表复制模型 ID，不要凭感觉手动输入。"),
        ("为什么工具连不上？", "请检查 Base URL 是否填写到 /v1，域名是否写错，网络是否能打开中转站后台。如果工具有代理设置，也请检查代理是否影响请求。"),
        ("Key 可以发给别人吗？", "不建议。Key 等同于你的额度使用凭证。别人拿到 Key 后产生的消耗，会计入你的账号。"),
    ]
    for question, answer in faqs:
        add_heading(doc, question, 2)
        add_para(doc, answer)

    add_heading(doc, "安全提醒", 1)
    for item in [
        "不要公开 API Key。",
        "不要把 Key 写进公开仓库。",
        "不要在截图里暴露 Key。",
        "定期查看用量记录，发现异常及时停用旧 Key。",
        "只在可信工具里填写 URL 和 Key。",
        "不要用平台进行违法、违规或侵犯他人权益的内容生成与调用。",
    ]:
        add_bullet(doc, item)

    add_heading(doc, "给客服或管理员的信息", 1)
    add_para(doc, "如果用户无法完成配置，可以让用户提供以下信息，方便排查。请提醒用户：截图时务必遮挡 API Key，只保留前后少量字符用于识别即可。")
    for item in [
        "注册账号或邮箱。",
        "兑换码截图或兑换提示截图。",
        "后台余额截图。",
        "工具里的 Base URL 截图。",
        "工具里的模型名截图。",
        "报错信息截图。",
    ]:
        add_number(doc, item)

    doc.save(OUT)


if __name__ == "__main__":
    build_doc()
    print(OUT)
