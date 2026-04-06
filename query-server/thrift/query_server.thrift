namespace py query_server

# 分词结果中的词
struct Word {
    1: string word,
    2: i32 count
}

# 分词响应
struct SegmentResult {
    1: string text,
    2: list<Word> words,
    3: i32 count
}

# 实体
struct Entity {
    1: string text,
    2: string type,
    3: i32 start_pos,
    4: i32 end_pos
}

# NER 响应
struct NERResult {
    1: string text,
    2: list<Entity> entities
}

# 综合分析结果
struct AnalysisResult {
    1: string text,
    2: optional list<Word> keywords,
    3: optional list<Entity> entities,
    4: optional list<double> embedding,
    5: optional i32 dimension
}

# 服务定义
service QueryServer {
    # 健康检查
    string health(),
    
    # 分词（提取重要词）
    list<SegmentResult> segment(1: list<string> texts, 2: i32 size),
    
    # NER 实体识别
    list<NERResult> ner(1: list<string> texts),
    
    # 向量化
    list<list<double>> embed(1: list<string> texts, 2: bool normalize),
    
    # 文本相似度
    double similarity(1: string text1, 2: string text2),
    
    # 综合分析
    list<AnalysisResult> analyze(
        1: list<string> texts,
        2: bool keywords,
        3: bool ner,
        4: bool embedding,
        5: i32 keyword_top_k
    )
}